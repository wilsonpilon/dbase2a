// Package msx provê ferramentas e geradores de código C/SDCC para o ecossistema MSX.
package msx

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ConvertIHXToBin converte o conteúdo de um arquivo Intel HEX (.ihx) gerado pelo SDCC
// em um array de bytes binário alinhado a partir do endereço base (padrão 0x0100 para MSX-DOS .COM).
func ConvertIHXToBin(ihxData string, baseAddr uint16) ([]byte, error) {
	scanner := bufio.NewScanner(strings.NewReader(ihxData))
	mem := make(map[uint32]byte)
	var maxAddr uint32 = 0
	var minAddr uint32 = 0xFFFFFFFF
	var extendedSegment uint32 = 0
	var extendedLinear uint32 = 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.HasPrefix(line, ":") {
			continue
		}

		// Linha Intel HEX: :LLAAAATTDD...CC
		// LL = Comprimento dos dados (1 byte)
		// AAAA = Endereço base (2 bytes)
		// TT = Tipo de registro (1 byte)
		// DD... = Dados
		// CC = Checksum (1 byte)
		lineContent := line[1:]
		if len(lineContent) < 10 {
			continue
		}

		rawBytes, err := hex.DecodeString(lineContent)
		if err != nil {
			return nil, fmt.Errorf("erro ao decodificar linha hex '%s': %w", line, err)
		}

		byteCount := int(rawBytes[0])
		offset := (uint32(rawBytes[1]) << 8) | uint32(rawBytes[2])
		recordType := rawBytes[3]

		if len(rawBytes) < 4+byteCount+1 {
			return nil, fmt.Errorf("linha hex incompleta: '%s'", line)
		}

		data := rawBytes[4 : 4+byteCount]

		switch recordType {
		case 0x00: // Dados
			addr := extendedSegment + extendedLinear + offset
			for i, b := range data {
				currAddr := addr + uint32(i)
				mem[currAddr] = b
				if currAddr > maxAddr {
					maxAddr = currAddr
				}
				if currAddr < minAddr {
					minAddr = currAddr
				}
			}
		case 0x01: // Fim de Arquivo (EOF)
			break
		case 0x02: // Endereço de Segmento Estendido
			if byteCount == 2 {
				seg := (uint32(data[0]) << 8) | uint32(data[1])
				extendedSegment = seg << 4
			}
		case 0x04: // Endereço Linear Estendido
			if byteCount == 2 {
				lin := (uint32(data[0]) << 8) | uint32(data[1])
				extendedLinear = lin << 16
			}
		default:
			// Outros tipos ignorados
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("erro ao ler dados hex: %w", err)
	}

	if len(mem) == 0 {
		return []byte{}, nil
	}

	start := uint32(baseAddr)
	if minAddr > start {
		// Se os dados começam após o baseAddr, mantemos baseAddr como ponto de partida
	} else if minAddr < start {
		start = minAddr
	}

	size := maxAddr - start + 1
	if size > 65536 {
		return nil, fmt.Errorf("tamanho do binário excede limite de memória de 64KB (%d bytes)", size)
	}

	bin := make([]byte, size)
	for addr, val := range mem {
		if addr >= start && addr <= maxAddr {
			bin[addr-start] = val
		}
	}

	return bin, nil
}

// ConvertIHXFileToCOM lê um arquivo .ihx e grava o binário executável .COM para MSX-DOS (base 0x0100).
func ConvertIHXFileToCOM(ihxFilePath, comFilePath string) error {
	data, err := os.ReadFile(ihxFilePath)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo IHX (%s): %w", ihxFilePath, err)
	}

	bin, err := ConvertIHXToBin(string(data), 0x0100)
	if err != nil {
		return fmt.Errorf("erro na conversão de IHX para COM: %w", err)
	}

	if err := os.WriteFile(comFilePath, bin, 0755); err != nil {
		return fmt.Errorf("erro ao salvar executável COM (%s): %w", comFilePath, err)
	}

	return nil
}

// FormatHexAddress formata um inteiro como endereço hexadecimal de 4 dígitos.
func FormatHexAddress(addr uint16) string {
	return strconv.FormatUint(uint64(addr), 16)
}

// DumpBinaryInfo gera um resumo legível do binário .COM.
func DumpBinaryInfo(bin []byte, baseAddr uint16) string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Tamanho do Executável: %d bytes (0x%04X - 0x%04X)\n", len(bin), baseAddr, baseAddr+uint16(len(bin))-1)
	return buf.String()
}
