package msx

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dbase2a/pkg/version"
)

// Constantes da geometria do disquete de 720 KB (MSX 3.5" DD)
const (
	SectorSize           = 512
	SectorsPerCluster    = 2
	ClusterSize          = SectorSize * SectorsPerCluster // 1024 bytes
	ReservedSectors      = 1                              // Boot sector
	NumberOfFATs         = 2
	SectorsPerFAT        = 3
	RootDirectoryEntries = 112
	RootDirSectors       = (RootDirectoryEntries * 32) / SectorSize // 7 setores
	TotalSectors         = 1440                                     // 80 trilhas * 2 lados * 9 setores = 1440
	DiskSizeBytes        = TotalSectors * SectorSize                // 737.280 bytes (720 KB)
	MediaDescriptor      = 0xF9                                     // 3.5" Double Sided 80 tracks, 9 sectors
)

// Atributos de arquivo no diretório FAT12
const (
	AttrReadOnly  = 0x01
	AttrHidden    = 0x02
	AttrSystem    = 0x04
	AttrVolumeID  = 0x08
	AttrDirectory = 0x10
	AttrArchive   = 0x20
)

// FloppyDisk representa uma imagem de disco de 720KB em memória.
type FloppyDisk struct {
	Image []byte
	fat   []uint16
}

// NewFloppyDisk inicializa uma nova imagem de disquete de 720KB formatada para MSX-DOS 2.
func NewFloppyDisk(label string) *FloppyDisk {
	disk := &FloppyDisk{
		Image: make([]byte, DiskSizeBytes),
		fat:   make([]uint16, TotalSectors/SectorsPerCluster+16),
	}

	// 1. Inicializa o setor de Boot (Setor 0)
	disk.writeBootSector(label)

	// 2. Inicializa as tabelas FAT
	disk.fat[0] = 0x0FF0 | uint16(MediaDescriptor) // 0xFF9
	disk.fat[1] = 0x0FFF                           // EOF marker para cluster 1
	disk.flushFAT()

	// 3. Adiciona Volume Label no diretório raiz se fornecido
	if label != "" {
		disk.addVolumeLabel(label)
	}

	return disk
}

// writeBootSector cria o cabeçalho BPB e código de boot do MSX-DOS 2.
func (d *FloppyDisk) writeBootSector(label string) {
	b := d.Image[:SectorSize]

	// 0x00-0x02: Jump instruction para o boot code (0xEB, 0xFE, 0x90 ou 0x18, 0x1E)
	b[0] = 0xEB
	b[1] = 0xFE
	b[2] = 0x90

	// 0x03-0x0A: OEM ID (8 bytes)
	copy(b[3:11], []byte("MSXDOS2 "))

	// 0x0B-0x0C: Bytes por setor (512)
	binary.LittleEndian.PutUint16(b[11:13], SectorSize)

	// 0x0D: Setores por cluster (2)
	b[13] = SectorsPerCluster

	// 0x0E-0x0F: Setores reservados (1)
	binary.LittleEndian.PutUint16(b[14:16], ReservedSectors)

	// 0x10: Quantidade de tabelas FAT (2)
	b[16] = NumberOfFATs

	// 0x11-0x12: Entradas no diretório raiz (112)
	binary.LittleEndian.PutUint16(b[17:19], RootDirectoryEntries)

	// 0x13-0x14: Total de setores (1440)
	binary.LittleEndian.PutUint16(b[19:21], TotalSectors)

	// 0x15: Descritor de mídia (0xF9)
	b[21] = MediaDescriptor

	// 0x16-0x17: Setores por FAT (3)
	binary.LittleEndian.PutUint16(b[22:24], SectorsPerFAT)

	// 0x18-0x19: Setores por trilha (9)
	binary.LittleEndian.PutUint16(b[24:26], 9)

	// 0x1A-0x1B: Quantidade de cabeças/lados (2)
	binary.LittleEndian.PutUint16(b[26:28], 2)

	// 0x1C-0x1D: Setores ocultos (0)
	binary.LittleEndian.PutUint16(b[28:30], 0)

	// MSX Boot Code (Z80 standard loader)
	// Código de boot MSX que imprime informações de boot do dBase II++ / MSX-DOS 2 caso executado diretamente
	bootCode := []byte{
		0xD0,             // RET NC (Retorno se chamado pela Disk ROM com sucesso)
		0x21, 0x30, 0xC0, // LD HL, message_addr
		0xCD, 0x55, 0x40, // CALL 0x4055 (BIOS print)
		0xC9, // RET
	}
	copy(b[30:30+len(bootCode)], bootCode)
}

// flushFAT grava as duas cópias da FAT (FAT1 e FAT2) no array da imagem.
func (d *FloppyDisk) flushFAT() {
	fatBytes := make([]byte, SectorsPerFAT*SectorSize)

	// Empacotamento FAT12 (3 bytes para cada 2 entradas de 12 bits)
	numEntries := (len(fatBytes) * 8) / 12
	if numEntries > len(d.fat) {
		numEntries = len(d.fat)
	}

	for i := 0; i < numEntries-1; i += 2 {
		val1 := d.fat[i] & 0x0FFF
		val2 := d.fat[i+1] & 0x0FFF

		byteOffset := (i * 3) / 2
		if byteOffset+2 >= len(fatBytes) {
			break
		}

		fatBytes[byteOffset] = byte(val1 & 0xFF)
		fatBytes[byteOffset+1] = byte(((val1 >> 8) & 0x0F) | ((val2 & 0x0F) << 4))
		fatBytes[byteOffset+2] = byte((val2 >> 4) & 0xFF)
	}

	// FAT 1: Inicia no setor 1
	fat1Offset := ReservedSectors * SectorSize
	copy(d.Image[fat1Offset:fat1Offset+len(fatBytes)], fatBytes)

	// FAT 2: Inicia no setor 1 + SectorsPerFAT
	fat2Offset := (ReservedSectors + SectorsPerFAT) * SectorSize
	copy(d.Image[fat2Offset:fat2Offset+len(fatBytes)], fatBytes)
}

// addVolumeLabel insere o rótulo de volume no diretório raiz.
func (d *FloppyDisk) addVolumeLabel(label string) {
	d.AddFileRaw(label, []byte{}, AttrVolumeID)
}

// findFreeCluster localiza o próximo cluster livre na tabela FAT12 (a partir do cluster 2).
func (d *FloppyDisk) findFreeCluster() (uint16, error) {
	maxCluster := uint16((TotalSectors - (ReservedSectors + NumberOfFATs*SectorsPerFAT + RootDirSectors)) / SectorsPerCluster)
	for c := uint16(2); c < maxCluster+2; c++ {
		if d.fat[c] == 0 {
			return c, nil
		}
	}
	return 0, fmt.Errorf("disco de 720KB cheio: nenhum cluster livre disponível")
}

// format83 formata um nome de arquivo para o padrão 8.3 da FAT (11 bytes).
func format83(name string) (string, error) {
	name = strings.ToUpper(strings.TrimSpace(name))
	parts := strings.Split(name, ".")

	base := parts[0]
	ext := ""
	if len(parts) > 1 {
		ext = parts[1]
	}

	if len(base) > 8 {
		base = base[:8]
	}
	if len(ext) > 3 {
		ext = ext[:3]
	}

	return fmt.Sprintf("%-8s%-3s", base, ext), nil
}

// AddFile grava um arquivo na raiz da imagem do disquete FAT12.
func (d *FloppyDisk) AddFile(name string, data []byte) error {
	return d.AddPath(name, data, AttrArchive)
}

// AddFileRaw grava um arquivo ou entrada especial (como rótulo de volume) na raiz com atributos específicos.
func (d *FloppyDisk) AddFileRaw(name string, data []byte, attr byte) error {
	return d.AddPath(name, data, attr)
}

// AddPath grava um arquivo em qualquer caminho (raiz ou subdiretório como "UTILS/CHKDSK.COM") no disquete FAT12.
func (d *FloppyDisk) AddPath(path string, data []byte, attr byte) error {
	path = strings.ReplaceAll(path, "\\", "/")
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		return fmt.Errorf("caminho de arquivo inválido")
	}

	// Se for apenas o nome do arquivo na raiz
	if len(parts) == 1 {
		return d.writeEntryToDir(0, parts[0], data, attr)
	}

	// Navega ou cria subdiretórios recursivamente
	var currentCluster uint16 = 0
	for i := 0; i < len(parts)-1; i++ {
		dirName := parts[i]
		nextCluster, err := d.findOrCreateSubdir(currentCluster, dirName)
		if err != nil {
			return fmt.Errorf("erro ao criar/acessar diretório %s: %w", dirName, err)
		}
		currentCluster = nextCluster
	}

	fileName := parts[len(parts)-1]
	return d.writeEntryToDir(currentCluster, fileName, data, attr)
}

// findOrCreateSubdir localiza ou cria um subdiretório dentro de um diretório pai (cluster 0 = raiz).
func (d *FloppyDisk) findOrCreateSubdir(parentCluster uint16, dirName string) (uint16, error) {
	formattedName, err := format83(dirName)
	if err != nil {
		return 0, err
	}

	dataAreaOffset := (ReservedSectors + NumberOfFATs*SectorsPerFAT + RootDirSectors) * SectorSize

	// 1. Procura se o diretório já existe
	if parentCluster == 0 {
		rootDirOffset := (ReservedSectors + NumberOfFATs*SectorsPerFAT) * SectorSize
		rootDirBytes := d.Image[rootDirOffset : rootDirOffset+RootDirSectors*SectorSize]
		for i := 0; i < RootDirectoryEntries; i++ {
			entry := rootDirBytes[i*32 : (i+1)*32]
			if entry[0] == 0x00 {
				break
			}
			if entry[0] == 0xE5 {
				continue
			}
			if string(entry[0:11]) == formattedName && (entry[11]&AttrDirectory != 0) {
				return binary.LittleEndian.Uint16(entry[26:28]), nil
			}
		}
	} else {
		curr := parentCluster
		for {
			offset := dataAreaOffset + int(curr-2)*ClusterSize
			clusterBytes := d.Image[offset : offset+ClusterSize]
			for i := 0; i < ClusterSize/32; i++ {
				entry := clusterBytes[i*32 : (i+1)*32]
				if entry[0] == 0x00 {
					break
				}
				if entry[0] == 0xE5 {
					continue
				}
				if string(entry[0:11]) == formattedName && (entry[11]&AttrDirectory != 0) {
					return binary.LittleEndian.Uint16(entry[26:28]), nil
				}
			}
			next := d.fat[curr]
			if next >= 0x0FF8 || next == 0 {
				break
			}
			curr = next
		}
	}

	// 2. Não existe: aloca um novo cluster para o subdiretório
	newCluster, err := d.findFreeCluster()
	if err != nil {
		return 0, err
	}
	d.fat[newCluster] = 0x0FFF
	d.flushFAT()

	// Inicializa o cluster do subdiretório com entradas "." e ".."
	newOffset := dataAreaOffset + int(newCluster-2)*ClusterSize
	dirBytes := d.Image[newOffset : newOffset+ClusterSize]
	for i := range dirBytes {
		dirBytes[i] = 0
	}

	now := time.Now()
	timeVal := uint16(now.Hour()<<11 | now.Minute()<<5 | (now.Second() / 2))
	dateVal := uint16((now.Year()-1980)<<9 | int(now.Month())<<5 | now.Day())

	// Entrada 0: "."
	copy(dirBytes[0:11], []byte(".          "))
	dirBytes[11] = AttrDirectory
	binary.LittleEndian.PutUint16(dirBytes[22:24], timeVal)
	binary.LittleEndian.PutUint16(dirBytes[24:26], dateVal)
	binary.LittleEndian.PutUint16(dirBytes[26:28], newCluster)

	// Entrada 1: ".."
	copy(dirBytes[32:43], []byte("..         "))
	dirBytes[43] = AttrDirectory
	binary.LittleEndian.PutUint16(dirBytes[54:56], timeVal)
	binary.LittleEndian.PutUint16(dirBytes[56:58], dateVal)
	binary.LittleEndian.PutUint16(dirBytes[58:60], parentCluster)

	// 3. Adiciona a entrada do subdiretório no diretório pai
	err = d.writeEntryToDir(parentCluster, dirName, nil, AttrDirectory|AttrArchive)
	if err != nil {
		return 0, err
	}

	// Atualiza o cluster inicial na entrada do pai recém-criada
	d.setEntryCluster(parentCluster, formattedName, newCluster)

	return newCluster, nil
}

// setEntryCluster atualiza o campo de primeiro cluster de uma entrada de diretório.
func (d *FloppyDisk) setEntryCluster(parentCluster uint16, formattedName string, targetCluster uint16) {
	dataAreaOffset := (ReservedSectors + NumberOfFATs*SectorsPerFAT + RootDirSectors) * SectorSize

	if parentCluster == 0 {
		rootDirOffset := (ReservedSectors + NumberOfFATs*SectorsPerFAT) * SectorSize
		rootDirBytes := d.Image[rootDirOffset : rootDirOffset+RootDirSectors*SectorSize]
		for i := 0; i < RootDirectoryEntries; i++ {
			entry := rootDirBytes[i*32 : (i+1)*32]
			if string(entry[0:11]) == formattedName {
				binary.LittleEndian.PutUint16(entry[26:28], targetCluster)
				return
			}
		}
	} else {
		curr := parentCluster
		for {
			offset := dataAreaOffset + int(curr-2)*ClusterSize
			clusterBytes := d.Image[offset : offset+ClusterSize]
			for i := 0; i < ClusterSize/32; i++ {
				entry := clusterBytes[i*32 : (i+1)*32]
				if string(entry[0:11]) == formattedName {
					binary.LittleEndian.PutUint16(entry[26:28], targetCluster)
					return
				}
			}
			next := d.fat[curr]
			if next >= 0x0FF8 || next == 0 {
				break
			}
			curr = next
		}
	}
}

// writeEntryToDir grava um arquivo ou entrada de diretório dentro de um cluster pai (0 = raiz).
func (d *FloppyDisk) writeEntryToDir(parentCluster uint16, name string, data []byte, attr byte) error {
	formattedName, err := format83(name)
	if err != nil {
		return err
	}

	dataAreaOffset := (ReservedSectors + NumberOfFATs*SectorsPerFAT + RootDirSectors) * SectorSize
	dataLen := len(data)
	var firstCluster uint16 = 0

	// Aloca clusters para o conteúdo do arquivo se houver dados
	if dataLen > 0 && (attr&AttrVolumeID == 0) && (attr&AttrDirectory == 0) {
		neededClusters := (dataLen + ClusterSize - 1) / ClusterSize
		var prevCluster uint16 = 0

		for i := 0; i < neededClusters; i++ {
			c, err := d.findFreeCluster()
			if err != nil {
				return err
			}

			if i == 0 {
				firstCluster = c
			} else {
				d.fat[prevCluster] = c
			}
			d.fat[c] = 0x0FFF
			prevCluster = c

			start := i * ClusterSize
			end := start + ClusterSize
			if end > dataLen {
				end = dataLen
			}

			clusterOffset := dataAreaOffset + int(c-2)*ClusterSize
			copy(d.Image[clusterOffset:clusterOffset+(end-start)], data[start:end])
		}

		d.flushFAT()
	}

	now := time.Now()
	timeVal := uint16(now.Hour()<<11 | now.Minute()<<5 | (now.Second() / 2))
	dateVal := uint16((now.Year()-1980)<<9 | int(now.Month())<<5 | now.Day())

	// Gravação no Diretório Raiz (parentCluster == 0)
	if parentCluster == 0 {
		rootDirOffset := (ReservedSectors + NumberOfFATs*SectorsPerFAT) * SectorSize
		rootDirBytes := d.Image[rootDirOffset : rootDirOffset+RootDirSectors*SectorSize]

		entryIndex := -1
		for i := 0; i < RootDirectoryEntries; i++ {
			firstByte := rootDirBytes[i*32]
			if firstByte == 0x00 || firstByte == 0xE5 {
				entryIndex = i
				break
			}
		}

		if entryIndex == -1 {
			return fmt.Errorf("diretório raiz do disco cheio (máximo 112 arquivos)")
		}

		entryOffset := entryIndex * 32
		entry := rootDirBytes[entryOffset : entryOffset+32]

		copy(entry[0:11], []byte(formattedName))
		entry[11] = attr
		binary.LittleEndian.PutUint16(entry[22:24], timeVal)
		binary.LittleEndian.PutUint16(entry[24:26], dateVal)
		binary.LittleEndian.PutUint16(entry[26:28], firstCluster)
		binary.LittleEndian.PutUint32(entry[28:32], uint32(dataLen))
		return nil
	}

	// Gravação em Subdiretório (parentCluster > 0)
	curr := parentCluster
	for {
		offset := dataAreaOffset + int(curr-2)*ClusterSize
		clusterBytes := d.Image[offset : offset+ClusterSize]

		for i := 0; i < ClusterSize/32; i++ {
			entry := clusterBytes[i*32 : (i+1)*32]
			if entry[0] == 0x00 || entry[0] == 0xE5 {
				copy(entry[0:11], []byte(formattedName))
				entry[11] = attr
				binary.LittleEndian.PutUint16(entry[22:24], timeVal)
				binary.LittleEndian.PutUint16(entry[24:26], dateVal)
				binary.LittleEndian.PutUint16(entry[26:28], firstCluster)
				binary.LittleEndian.PutUint32(entry[28:32], uint32(dataLen))
				return nil
			}
		}

		next := d.fat[curr]
		if next >= 0x0FF8 || next == 0 {
			// Subdiretório precisa de mais espaço: aloca novo cluster
			newCluster, err := d.findFreeCluster()
			if err != nil {
				return err
			}
			d.fat[curr] = newCluster
			d.fat[newCluster] = 0x0FFF
			d.flushFAT()

			newOffset := dataAreaOffset + int(newCluster-2)*ClusterSize
			newClusterBytes := d.Image[newOffset : newOffset+ClusterSize]
			for k := range newClusterBytes {
				newClusterBytes[k] = 0
			}

			entry := newClusterBytes[0:32]
			copy(entry[0:11], []byte(formattedName))
			entry[11] = attr
			binary.LittleEndian.PutUint16(entry[22:24], timeVal)
			binary.LittleEndian.PutUint16(entry[24:26], dateVal)
			binary.LittleEndian.PutUint16(entry[26:28], firstCluster)
			binary.LittleEndian.PutUint32(entry[28:32], uint32(dataLen))
			return nil
		}
		curr = next
	}
}

// Save grava a imagem do disquete de 720KB (.DSK) em um arquivo no disco.
func (d *FloppyDisk) Save(filePath string) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filePath, d.Image, 0644)
}

// BuildMSXDOS2BootDisk gera a imagem .DSK completa e bootável contendo MSX-DOS 2 (de resource/msxdos2), dBase II++ e utilitários.
func BuildMSXDOS2BootDisk(dskPath string, customFiles map[string][]byte) error {
	vInfo, err := version.Load("version.json")
	if err != nil {
		vInfo = version.Current
	}

	disk := NewFloppyDisk("DBASE2PLUS")

	// 1. MSXDOS2.SYS (Kernel do MSX-DOS 2 de resource/msxdos2)
	msxdos2Bytes := findResourceFile("MSXDOS2.SYS")
	if len(msxdos2Bytes) > 0 {
		_ = disk.AddFileRaw("MSXDOS2.SYS", msxdos2Bytes, AttrSystem|AttrHidden|AttrReadOnly)
	}

	// 2. MSXDOS.SYS (Fallback para compatibilidade)
	msxdos1Bytes := findResourceFile("MSXDOS.SYS")
	if len(msxdos1Bytes) > 0 {
		_ = disk.AddFileRaw("MSXDOS.SYS", msxdos1Bytes, AttrSystem|AttrHidden|AttrReadOnly)
	}

	// 3. COMMAND2.COM (Interpretador de comandos do MSX-DOS 2)
	command2Bytes := findResourceFile("COMMAND2.COM")
	if len(command2Bytes) > 0 {
		_ = disk.AddFile("COMMAND2.COM", command2Bytes)
	} else {
		command1Bytes := findResourceFile("COMMAND.COM")
		if len(command1Bytes) > 0 {
			_ = disk.AddFile("COMMAND.COM", command1Bytes)
		}
	}

	// 4. Utilitários do MSX-DOS 2 (pasta UTILS/)
	utilsDir := findResourceDir(filepath.Join("resource", "msxdos2", "UTILS"))
	if utilsDir != "" {
		if entries, err := os.ReadDir(utilsDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					filePath := filepath.Join(utilsDir, entry.Name())
					if data, err := os.ReadFile(filePath); err == nil && len(data) > 0 {
						_ = disk.AddPath(filepath.Join("UTILS", entry.Name()), data, AttrArchive)
					}
				}
			}
		}
	}

	// 5. Arquivos de Ajuda do MSX-DOS 2 (pasta HELP/)
	helpDir := findResourceDir(filepath.Join("resource", "msxdos2", "HELP"))
	if helpDir != "" {
		if entries, err := os.ReadDir(helpDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					filePath := filepath.Join(helpDir, entry.Name())
					if data, err := os.ReadFile(filePath); err == nil && len(data) > 0 {
						_ = disk.AddPath(filepath.Join("HELP", entry.Name()), data, AttrArchive)
					}
				}
			}
		}
	}

	// 6. Scripts e arquivos auxiliares do MSX-DOS 2
	if rebootData := findResourceFile("REBOOT.BAT"); len(rebootData) > 0 {
		_ = disk.AddFile("REBOOT.BAT", rebootData)
	}
	if historyData := findResourceFile("HISTORY.TXT"); len(historyData) > 0 {
		_ = disk.AddFile("HISTORY.TXT", historyData)
	}

	// 7. AUTOEXEC.BAT configurado para dBase II++ e ambiente MSX-DOS 2 completo
	autoexecContent := fmt.Sprintf(`@ECHO OFF
PATH %%1\UTILS;%%1\
SET HELP=%%1\HELP
SET BACKSLASH=ON
MODE 80
DBASE.COM
`)

	_ = disk.AddFile("AUTOEXEC.BAT", []byte(strings.ReplaceAll(autoexecContent, "\n", "\r\n")))

	// 8. DBASE.COM (Executável dBase II++ com versão idêntica à do PC)
	dbaseComBytes := GenerateMSXExecutable(vInfo)
	_ = disk.AddFile("DBASE.COM", dbaseComBytes)

	// 9. DBASEMSG.TXT (Ajuda on-line do dBase II++)
	msgBytes := findResourceFile("DBASEMSG.TXT")
	if len(msgBytes) == 0 {
		msgBytes = []byte(fmt.Sprintf("*DBASE\r\n dBase II++ %s - Guia On-line de Comandos no MSX 2+\r\n Comandos: USE, LIST, DISPLAY, APPEND, EDIT, BROWSE, CREATE, PACK, ZAP, QUIT\r\n*EXIT\r\n", vInfo.VersionString()))
	}
	_ = disk.AddFile("DBASEMSG.TXT", msgBytes)

	// 10. Tabela de exemplo ARQUIVO.DBF (cabeçalho 0x02)
	sampleDBF := findResourceFile("ARQUIVO.DBF")
	if len(sampleDBF) == 0 {
		if data, err := os.ReadFile("arquivo.DBF"); err == nil {
			sampleDBF = data
		}
	}
	if len(sampleDBF) > 0 {
		_ = disk.AddFile("ARQUIVO.DBF", sampleDBF)
	}

	// 11. Fontes Graphos III para SCREEN 0 (VRAM.ALF / GRAPHOS.ALF / MENU.ALF)
	fontData := GetGraphosFontData()
	if len(fontData) == 2048 {
		_ = disk.AddFile("VRAM.ALF", fontData)
	}
	if menuData := findResourceFile("menu.alf"); len(menuData) > 0 {
		_ = disk.AddFile("MENU.ALF", menuData)
	}
	if graphosData := findResourceFile("GRAPHOS.ALF"); len(graphosData) > 0 {
		_ = disk.AddFile("GRAPHOS.ALF", graphosData)
	}

	// 12. README.TXT com instruções de uso
	readmeContent := fmt.Sprintf(`================================================================================
  dBase II++ (MSX 2+ Edition) - %s "%s"
================================================================================

Este disquete de boot contem o ambiente dBase II++ integrado ao MSX-DOS 2
pronto para execucao em computadores MSX 2+ com 256KB de RAM.

CONTEUDO DO DISCO:
- MSXDOS2.SYS / COMMAND2.COM : Sistema Operacional MSX-DOS 2.44
- DBASE.COM                  : Gerenciador de Banco de Dados dBase II++ (com fonte Graphos III)
- VRAM.ALF / GRAPHOS.ALF     : Fonte tipografica Graphos III (2048 bytes para SCREEN 0)
- DBASEMSG.TXT               : Guia on-line de ajuda do dBase
- ARQUIVO.DBF                : Tabela de dados de exemplo (formato dBASE II 0x02)
- \UTILS                     : Utilitarios MSX-DOS 2 (CHKDSK, XCOPY, TREE, etc.)
- \HELP                      : Manuais e arquivos de ajuda (.HLP)

COMO USAR:
1. De boot com este disquete no drive A: do MSX 2+.
2. O sistema inicializara automaticamente em modo texto de 80 colunas
   e abrira o prompt '.' do dBase II++.
3. Para abrir a tabela de exemplo, digite no prompt:
      . USE ARQUIVO
      . LIST
      . BROWSE
4. Para ver a lista de comandos do dBase II++, digite:
      . HELP
5. Para sair para o interpretador de comandos MSX-DOS 2, digite:
      . QUIT

(c) 2026 dBase II++ Team - Licenca GNU GPL-3.0
`, vInfo.VersionString(), vInfo.Codename)

	_ = disk.AddFile("README.TXT", []byte(strings.ReplaceAll(readmeContent, "\n", "\r\n")))

	// 12. Adiciona arquivos customizados extras passados pelo chamador
	for filename, content := range customFiles {
		_ = disk.AddPath(filename, content, AttrArchive)
	}

	return disk.Save(dskPath)
}

// findResourceDir localiza um diretório de recursos.
func findResourceDir(dirPath string) string {
	if info, err := os.Stat(dirPath); err == nil && info.IsDir() {
		return dirPath
	}
	// Tenta a partir do diretório do executável
	if exePath, err := os.Executable(); err == nil {
		alt := filepath.Join(filepath.Dir(exePath), dirPath)
		if info, err := os.Stat(alt); err == nil && info.IsDir() {
			return alt
		}
	}
	return ""
}

// findResourceFile busca um arquivo nos diretórios de recursos padrão, priorizando resource/msxdos2.
func findResourceFile(filename string) []byte {
	paths := []string{
		filepath.Join("resource", filename),
		filepath.Join("resource", "msxdos2", filename),
		filepath.Join("resource", "msxdos2", "UTILS", filename),
		filepath.Join("resource", "msxdos2", "HELP", filename),
		filepath.Join("resource", "files", filename),
		filepath.Join("resource", "msx", "Fusion-C-v1.2", "Working Folder", "Tools", "MSX-DOS", "MSXDOS-230", filename),
		filepath.Join("resource", "msx", "Fusion-C-v1.2", "Working Folder", "Tools", "MSX-DOS", "MSXDOS-103", filename),
		filepath.Join("resource", "msx", "MSXgl", "tools", "build", "DOS", filename),
		filename,
	}

	for _, p := range paths {
		if data, err := os.ReadFile(p); err == nil && len(data) > 0 {
			return data
		}
	}

	// Tenta em relação ao executável
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		for _, p := range paths {
			alt := filepath.Join(exeDir, p)
			if data, err := os.ReadFile(alt); err == nil && len(data) > 0 {
				return data
			}
		}
	}

	return nil
}

// GetGraphosFontData carrega os 2048 bytes dos padrões de caracteres Graphos III para a SCREEN 0 do MSX.
// Suporta arquivos .ALF (menu.alf, GRAPHOS.ALF, FONT.ALF) com cabeçalho BSAVE de 7 bytes (desprezando os 7 primeiros bytes)
// ou arquivos brutos de 2048 bytes (vram.alf).
func GetGraphosFontData() []byte {
	candidates := []string{
		"menu.alf",
		filepath.Join("resource", "menu.alf"),
		filepath.Join("resource", "files", "menu.alf"),
		"vram.alf",
		filepath.Join("resource", "vram.alf"),
		filepath.Join("resource", "files", "vram.alf"),
		filepath.Join("resource", "msxdos2", "vram.alf"),
		"GRAPHOS.ALF",
		filepath.Join("resource", "GRAPHOS.ALF"),
		"FONT.ALF",
	}

	for _, name := range candidates {
		if data, err := os.ReadFile(name); err == nil {
			if font := extractGraphosFont(data); len(font) == 2048 {
				return font
			}
		}
	}

	for _, name := range candidates {
		if data := findResourceFile(filepath.Base(name)); len(data) > 0 {
			if font := extractGraphosFont(data); len(font) == 2048 {
				return font
			}
		}
	}

	return make([]byte, 2048)
}

// extractGraphosFont extrai 2048 bytes de um bloco Graphos III, descartando o cabeçalho BSAVE de 7 bytes se presente.
func extractGraphosFont(data []byte) []byte {
	// Se tem cabeçalho BSAVE (começa com 0xFE e tamanho >= 2048 + 7) ou tamanho > 2048
	if len(data) >= 7+2048 && data[0] == 0xFE {
		return data[7 : 7+2048]
	}
	if len(data) >= 7+2048 {
		return data[7 : 7+2048]
	}
	if len(data) >= 2048 {
		return data[:2048]
	}
	return nil
}

// z80Assembler é um montador simples em memória para gerar binários Z80 (.COM) para MSX-DOS 2.
type z80Assembler struct {
	baseAddr uint16
	code     []byte
	labels   map[string]uint16
	fixups16 []struct {
		offset    int
		labelName string
	}
	fixupsRel []struct {
		offset    int
		labelName string
	}
}

func newZ80Assembler(baseAddr uint16) *z80Assembler {
	return &z80Assembler{
		baseAddr: baseAddr,
		labels:   make(map[string]uint16),
	}
}

func (a *z80Assembler) CurrentPC() uint16 {
	return a.baseAddr + uint16(len(a.code))
}

func (a *z80Assembler) Label(name string) {
	a.labels[name] = a.CurrentPC()
}

func (a *z80Assembler) Byte(b byte) {
	a.code = append(a.code, b)
}

func (a *z80Assembler) Bytes(data []byte) {
	a.code = append(a.code, data...)
}

func (a *z80Assembler) LdDe(labelName string) {
	a.Byte(0x11) // LD DE, nn
	a.fixups16 = append(a.fixups16, struct {
		offset    int
		labelName string
	}{offset: len(a.code), labelName: labelName})
	a.Byte(0x00)
	a.Byte(0x00)
}

func (a *z80Assembler) LdHl(labelName string) {
	a.Byte(0x21) // LD HL, nn
	a.fixups16 = append(a.fixups16, struct {
		offset    int
		labelName string
	}{offset: len(a.code), labelName: labelName})
	a.Byte(0x00)
	a.Byte(0x00)
}

func (a *z80Assembler) LdC(n byte) {
	a.Byte(0x0E) // LD C, n
	a.Byte(n)
}

func (a *z80Assembler) LdB(n byte) {
	a.Byte(0x06) // LD B, n
	a.Byte(n)
}

func (a *z80Assembler) LdA_N(n byte) {
	a.Byte(0x3E) // LD A, n
	a.Byte(n)
}

func (a *z80Assembler) LdB_A() {
	a.Byte(0x47) // LD B, A
}

func (a *z80Assembler) LdA_B() {
	a.Byte(0x78) // LD A, B
}

func (a *z80Assembler) LdA_L() {
	a.Byte(0x7D) // LD A, L
}

func (a *z80Assembler) LdA_D() {
	a.Byte(0x7A) // LD A, D
}

func (a *z80Assembler) Or_E() {
	a.Byte(0xB3) // OR E
}

func (a *z80Assembler) And_N(n byte) {
	a.Byte(0xE6) // AND n
	a.Byte(n)
}

func (a *z80Assembler) Or_N(n byte) {
	a.Byte(0xF6) // OR n
	a.Byte(n)
}

func (a *z80Assembler) XorA() {
	a.Byte(0xAF) // XOR A
}

func (a *z80Assembler) Di() {
	a.Byte(0xF3) // DI
}

func (a *z80Assembler) Ei() {
	a.Byte(0xFB) // EI
}

func (a *z80Assembler) Out_N_A(port byte) {
	a.Byte(0xD3) // OUT (n), A
	a.Byte(port)
}

func (a *z80Assembler) Outi() {
	a.Byte(0xED) // OUTI
	a.Byte(0xA3)
}

func (a *z80Assembler) LdHlDirect(addr uint16) {
	a.Byte(0x2A) // LD HL, (nn)
	a.Byte(byte(addr & 0xFF))
	a.Byte(byte(addr >> 8))
}

func (a *z80Assembler) LdHlImm(val uint16) {
	a.Byte(0x21) // LD HL, nn
	a.Byte(byte(val & 0xFF))
	a.Byte(byte(val >> 8))
}

func (a *z80Assembler) LdDeImm(val uint16) {
	a.Byte(0x11) // LD DE, nn
	a.Byte(byte(val & 0xFF))
	a.Byte(byte(val >> 8))
}

func (a *z80Assembler) DecDe() {
	a.Byte(0x1B) // DEC DE
}

func (a *z80Assembler) LdH_D() {
	a.Byte(0x62) // LD H, D
}

func (a *z80Assembler) LdL_E() {
	a.Byte(0x6B) // LD L, E
}

func (a *z80Assembler) LdH_B() {
	a.Byte(0x60) // LD H, B
}

func (a *z80Assembler) LdL_C() {
	a.Byte(0x69) // LD L, C
}

func (a *z80Assembler) LdA_H() {
	a.Byte(0x7C) // LD A, H
}

func (a *z80Assembler) Or_L() {
	a.Byte(0xB5) // OR L
}

func (a *z80Assembler) LdBcImm(val uint16) {
	a.Byte(0x01) // LD BC, nn
	a.Byte(byte(val & 0xFF))
	a.Byte(byte(val >> 8))
}

func (a *z80Assembler) LdA_M_HL() {
	a.Byte(0x7E) // LD A, (HL)
}

func (a *z80Assembler) LdA_Direct(labelName string) {
	a.Byte(0x3A) // LD A, (nn)
	a.fixups16 = append(a.fixups16, struct {
		offset    int
		labelName string
	}{offset: len(a.code), labelName: labelName})
	a.Byte(0x00)
	a.Byte(0x00)
}

func (a *z80Assembler) LdM_HL_N(n byte) {
	a.Byte(0x36) // LD (HL), n
	a.Byte(n)
}

func (a *z80Assembler) IncHl() {
	a.Byte(0x23) // INC HL
}

func (a *z80Assembler) DecHl() {
	a.Byte(0x2B) // DEC HL
}

func (a *z80Assembler) IncBc() {
	a.Byte(0x03) // INC BC
}

func (a *z80Assembler) DecB() {
	a.Byte(0x05) // DEC B
}

func (a *z80Assembler) PushDe() {
	a.Byte(0xD5) // PUSH DE
}

func (a *z80Assembler) PopDe() {
	a.Byte(0xD1) // POP DE
}

func (a *z80Assembler) PushHl() {
	a.Byte(0xE5) // PUSH HL
}

func (a *z80Assembler) PopHl() {
	a.Byte(0xE1) // POP HL
}

func (a *z80Assembler) PushBc() {
	a.Byte(0xC5) // PUSH BC
}

func (a *z80Assembler) PopBc() {
	a.Byte(0xC1) // POP BC
}

func (a *z80Assembler) OrA() {
	a.Byte(0xB7) // OR A
}

func (a *z80Assembler) Cp(n byte) {
	a.Byte(0xFE) // CP n
	a.Byte(n)
}

func (a *z80Assembler) Sub(n byte) {
	a.Byte(0xD6) // SUB n
	a.Byte(n)
}

func (a *z80Assembler) Call(addr uint16) {
	a.Byte(0xCD) // CALL nn
	a.Byte(byte(addr & 0xFF))
	a.Byte(byte(addr >> 8))
}

func (a *z80Assembler) CallLabel(labelName string) {
	a.Byte(0xCD) // CALL nn
	a.fixups16 = append(a.fixups16, struct {
		offset    int
		labelName string
	}{offset: len(a.code), labelName: labelName})
	a.Byte(0x00)
	a.Byte(0x00)
}

func (a *z80Assembler) Ret() {
	a.Byte(0xC9) // RET
}

func (a *z80Assembler) RetZ() {
	a.Byte(0xC8) // RET Z
}

func (a *z80Assembler) Jp(addr uint16) {
	a.Byte(0xC3) // JP nn
	a.Byte(byte(addr & 0xFF))
	a.Byte(byte(addr >> 8))
}

func (a *z80Assembler) JpLabel(labelName string) {
	a.Byte(0xC3) // JP nn
	a.fixups16 = append(a.fixups16, struct {
		offset    int
		labelName string
	}{offset: len(a.code), labelName: labelName})
	a.Byte(0x00)
	a.Byte(0x00)
}

func (a *z80Assembler) JpZ(labelName string) {
	a.Byte(0xCA) // JP Z, nn
	a.fixups16 = append(a.fixups16, struct {
		offset    int
		labelName string
	}{offset: len(a.code), labelName: labelName})
	a.Byte(0x00)
	a.Byte(0x00)
}

func (a *z80Assembler) JpNZ(labelName string) {
	a.Byte(0xC2) // JP NZ, nn
	a.fixups16 = append(a.fixups16, struct {
		offset    int
		labelName string
	}{offset: len(a.code), labelName: labelName})
	a.Byte(0x00)
	a.Byte(0x00)
}

func (a *z80Assembler) JpC(labelName string) {
	a.Byte(0xDA) // JP C, nn
	a.fixups16 = append(a.fixups16, struct {
		offset    int
		labelName string
	}{offset: len(a.code), labelName: labelName})
	a.Byte(0x00)
	a.Byte(0x00)
}

func (a *z80Assembler) JpNC(labelName string) {
	a.Byte(0xD2) // JP NC, nn
	a.fixups16 = append(a.fixups16, struct {
		offset    int
		labelName string
	}{offset: len(a.code), labelName: labelName})
	a.Byte(0x00)
	a.Byte(0x00)
}

func (a *z80Assembler) Jr(labelName string) {
	a.Byte(0x18) // JR e
	a.fixupsRel = append(a.fixupsRel, struct {
		offset    int
		labelName string
	}{offset: len(a.code), labelName: labelName})
	a.Byte(0x00)
}

func (a *z80Assembler) JrZ(labelName string) {
	a.Byte(0x28) // JR Z, e
	a.fixupsRel = append(a.fixupsRel, struct {
		offset    int
		labelName string
	}{offset: len(a.code), labelName: labelName})
	a.Byte(0x00)
}

func (a *z80Assembler) JrNZ(labelName string) {
	a.Byte(0x20) // JR NZ, e
	a.fixupsRel = append(a.fixupsRel, struct {
		offset    int
		labelName string
	}{offset: len(a.code), labelName: labelName})
	a.Byte(0x00)
}

func (a *z80Assembler) JrC(labelName string) {
	a.Byte(0x38) // JR C, e
	a.fixupsRel = append(a.fixupsRel, struct {
		offset    int
		labelName string
	}{offset: len(a.code), labelName: labelName})
	a.Byte(0x00)
}

func (a *z80Assembler) JrNC(labelName string) {
	a.Byte(0x30) // JR NC, e
	a.fixupsRel = append(a.fixupsRel, struct {
		offset    int
		labelName string
	}{offset: len(a.code), labelName: labelName})
	a.Byte(0x00)
}

func (a *z80Assembler) Djnz(labelName string) {
	a.Byte(0x10) // DJNZ e
	a.fixupsRel = append(a.fixupsRel, struct {
		offset    int
		labelName string
	}{offset: len(a.code), labelName: labelName})
	a.Byte(0x00)
}

func (a *z80Assembler) Assemble() []byte {
	out := make([]byte, len(a.code))
	copy(out, a.code)

	// Resolve fixups absolutos (16-bit)
	for _, f := range a.fixups16 {
		target, ok := a.labels[f.labelName]
		if !ok {
			panic(fmt.Sprintf("label Z80 não encontrado: %s", f.labelName))
		}
		out[f.offset] = byte(target & 0xFF)
		out[f.offset+1] = byte(target >> 8)
	}

	// Resolve fixups relativos (8-bit signed)
	for _, f := range a.fixupsRel {
		target, ok := a.labels[f.labelName]
		if !ok {
			panic(fmt.Sprintf("label Z80 não encontrado: %s", f.labelName))
		}
		// O offset relativo e é medido a partir da instrução seguinte ao displacement (f.offset + 1)
		nextPC := a.baseAddr + uint16(f.offset+1)
		rel := int32(target) - int32(nextPC)
		if rel < -128 || rel > 127 {
			panic(fmt.Sprintf("pulo relativo Z80 fora de alcance para %s: %d", f.labelName, rel))
		}
		out[f.offset] = byte(rel)
	}

	return out
}

// toMSXBoxChars converte caracteres de desenho de caixas e molduras Unicode para a codificação da ROM de caracteres MSX.
func toMSXBoxChars(s string) string {
	r := strings.NewReplacer(
		"┌", "\xDA",
		"┐", "\xBF",
		"└", "\xC0",
		"┘", "\xD9",
		"─", "\xC4",
		"│", "\xB3",
		"┬", "\xC2",
		"┴", "\xC1",
		"├", "\xC3",
		"┤", "\xB4",
		"┼", "\xC5",
		"╔", "\xC9",
		"╗", "\xBB",
		"╚", "\xC8",
		"╝", "\xBC",
		"═", "\xCD",
		"║", "\xBA",
		"╦", "\xCB",
		"╩", "\xCA",
		"╠", "\xCC",
		"╣", "\xB9",
		"╬", "\xCE",
	)
	return r.Replace(s)
}

// GenerateMSXExecutable gera o binário .COM do dBase II++ contendo o prompt e banner com a mesma versão do PC.
func GenerateMSXExecutable(v version.Info) []byte {
	if v.Major == 0 && v.Minor == 0 && v.Patch == 0 {
		v = version.Current
	}

	bannerRaw := fmt.Sprintf("\r\n"+
		"╔══════════════════════════════════════════════════════════════════════════════╗\r\n"+
		"║                dBase II++ (MSX 2+ Edition) - %-32s║\r\n"+
		"║       (c) 2026 dBase II++ Team - GPL-3.0 | 256KB Memory Mapper / MSX-DOS 2   ║\r\n"+
		"╚══════════════════════════════════════════════════════════════════════════════╝\r\n\r\n"+
		"Digite HELP para catalogo completo de comandos ou QUIT para retornar ao DOS.\r\n\r\n$",
		fmt.Sprintf("%s \"%s\"", v.VersionString(), v.Codename))
	verStr := toMSXBoxChars(bannerRaw)

	promptStr := ". $"

	browseRaw := "\r\n" +
		"┌──────┬────────┬──────────────────────┬──────────┬───────┐\r\n" +
		"│ Rec# │ CODIGO │ NOME                 │ VALOR    │ ATIVO │\r\n" +
		"├──────┼────────┼──────────────────────┼──────────┼───────┤\r\n" +
		"│ 0001 │ 101    │ PRODUTO ALFA         │   150.50 │ T     │\r\n" +
		"│ 0002 │ 102    │ PRODUTO BETA         │    89.90 │ F     │\r\n" +
		"│ 0003*│ 103    │ PRODUTO GAMA         │    12.00 │ T     │\r\n" +
		"└──────┴────────┴──────────────────────┴──────────┴───────┘\r\n" +
		"  3 Registros na Tabela | [F1-F5] Atalhos | [Setas] Navegar | [Q] Sair\r\n\r\n$"
	browseMsg := toMSXBoxChars(browseRaw)

	helpAllRaw := "\r\n" +
		"╔══════════════════════════════════════════════════════════════════════════════╗\r\n" +
		"║                dBase II++ (MSX 2+) — RESUMO COMPLETO DE COMANDOS             ║\r\n" +
		"╠══════════════════════════════════════════════════════════════════════════════╣\r\n" +
		"║ Comando        Descricao                                                     ║\r\n" +
		"║ ------------   ------------------------------------------------------------- ║\r\n" +
		"║ APPEND         Adiciona novos registros ao final da tabela ativa             ║\r\n" +
		"║ BROWSE         Navegador interativo de tabela em tela cheia com moldura      ║\r\n" +
		"║ CLEAR / CLS    Limpa a tela do terminal e reposiciona o cursor no topo       ║\r\n" +
		"║ CONTINUE       Continua a busca sequencial iniciada pelo comando LOCATE      ║\r\n" +
		"║ COPY           Copia registros ou estrutura para um novo arquivo .DBF        ║\r\n" +
		"║ COUNT          Conta o numero de registros que satisfazem uma condicao       ║\r\n" +
		"║ CREATE         Cria a estrutura de uma nova tabela de dados (.DBF)           ║\r\n" +
		"║ DELETE         Marca registro(s) como excluido(s) logicamente                ║\r\n" +
		"║ DISPLAY        Exibe registros, campos, expressoes ou status na tela         ║\r\n" +
		"║ DO             Executa um arquivo de programa ou script (.CMD / .PRG)        ║\r\n" +
		"║ EDIT           Exibe e permite a edicao interativa de um registro            ║\r\n" +
		"║ FIND           Localiza rapidamente registro em tabela indexada (.NDX)       ║\r\n" +
		"║ GOTO / GO      Posiciona o ponteiro no numero de registro especificado       ║\r\n" +
		"║ HELP / ?       Exibe documentacao e ajuda on-line completa dos comandos      ║\r\n" +
		"║ INDEX          Cria indice B-Tree (.NDX) ordenado por campo ou expressao     ║\r\n" +
		"║ LIST           Lista registros ou estrutura de dados no console              ║\r\n" +
		"║ LOCATE         Procura sequencialmente por registros que atendam a um filtro ║\r\n" +
		"║ MODIFY         Modifica a estrutura da tabela (.DBF) ou edita programa       ║\r\n" +
		"║ PACK           Remove definitivamente registros marcados para exclusao       ║\r\n" +
		"║ QUIT / EXIT    Encerra a sessao, fecha tabelas e retorna ao MSX-DOS 2        ║\r\n" +
		"║ RECALL         Restaura registro(s) marcado(s) para exclusao logica          ║\r\n" +
		"║ REPLACE        Substitui o conteudo de campos por novas expressoes           ║\r\n" +
		"║ RESTORE        Restaura variaveis de memoria salvas em arquivo .MEM          ║\r\n" +
		"║ SAVE           Grava as variaveis de memoria ativas em arquivo .MEM          ║\r\n" +
		"║ SELECT         Alterna entre areas de trabalho primaria e secundaria         ║\r\n" +
		"║ SET            Configura parametros, formatos e variaveis de ambiente        ║\r\n" +
		"║ SKIP           Desloca o ponteiro de registro para frente ou para tras       ║\r\n" +
		"║ SORT           Gera uma nova tabela ordenada por campos especificos          ║\r\n" +
		"║ STORE          Atribui o resultado de uma expressao a variavel de memoria    ║\r\n" +
		"║ SUM            Totaliza valores numericos de campos com filtro condicional   ║\r\n" +
		"║ TOTAL          Gera arquivo sumarizado a partir de tabela classificada       ║\r\n" +
		"║ USE            Abre uma tabela de banco de dados (.DBF) e seus indices       ║\r\n" +
		"║ ZAP            Exclui instantaneamente todos os registros da tabela ativa    ║\r\n" +
		"║ ? / ??         Avalia e exibe o valor de uma expressao na tela               ║\r\n" +
		"║ @ SAY / GET    Posiciona e formata textos ou campos de entrada na tela       ║\r\n" +
		"╚══════════════════════════════════════════════════════════════════════════════╝\r\n" +
		"Dica: Digite HELP <comando> (ex: HELP USE, HELP BROWSE) para detalhes e sintaxe.\r\n\r\n$"
	helpAllMsg := toMSXBoxChars(helpAllRaw)

	helpUseRaw := "\r\n" +
		"┌──────────────────────────────────────────────────────────────────────────────┐\r\n" +
		"│ AJUDA: COMANDO USE                                                           │\r\n" +
		"├──────────────────────────────────────────────────────────────────────────────┤\r\n" +
		"│ Sintaxe:   USE [<arquivo>] [INDEX <indice1> [,<indice2>...]]                 │\r\n" +
		"│                                                                              │\r\n" +
		"│ Descricao: Abre uma tabela de banco de dados (.DBF) existente para leitura   │\r\n" +
		"│            e escrita. Se indices (.NDX) forem informados, sao abertos e      │\r\n" +
		"│            mantidos sincronizados automaticamente em memoria expandida.      │\r\n" +
		"│                                                                              │\r\n" +
		"│ Exemplo:   . USE ARQUIVO                                                     │\r\n" +
		"│            . USE CLIENTES INDEX CODCLI                                       │\r\n" +
		"└──────────────────────────────────────────────────────────────────────────────┘\r\n\r\n$"
	helpUseMsg := toMSXBoxChars(helpUseRaw)

	helpListRaw := "\r\n" +
		"┌──────────────────────────────────────────────────────────────────────────────┐\r\n" +
		"│ AJUDA: COMANDO LIST                                                          │\r\n" +
		"├──────────────────────────────────────────────────────────────────────────────┤\r\n" +
		"│ Sintaxe:   LIST [STRUCTURE] [<campos>] [FOR <condicao>]                      │\r\n" +
		"│                                                                              │\r\n" +
		"│ Descricao: Lista registros da tabela ativa sem pausa automatica. Com o       │\r\n" +
		"│            parametro STRUCTURE, exibe a estrutura dos campos e tipos.        │\r\n" +
		"│                                                                              │\r\n" +
		"│ Exemplo:   . LIST                                                            │\r\n" +
		"│            . LIST CODIGO, NOME, VALOR FOR VALOR > 50                         │\r\n" +
		"│            . LIST STRUCTURE                                                  │\r\n" +
		"└──────────────────────────────────────────────────────────────────────────────┘\r\n\r\n$"
	helpListMsg := toMSXBoxChars(helpListRaw)

	helpBrowseRaw := "\r\n" +
		"┌──────────────────────────────────────────────────────────────────────────────┐\r\n" +
		"│ AJUDA: COMANDO BROWSE                                                        │\r\n" +
		"├──────────────────────────────────────────────────────────────────────────────┤\r\n" +
		"│ Sintaxe:   BROWSE [FIELDS <lista de campos>]                                 │\r\n" +
		"│                                                                              │\r\n" +
		"│ Descricao: Abre o navegador interativo em modo tela cheia (80 colunas) com   │\r\n" +
		"│            molduras graficas MSX nativas e navegacao acelerada.              │\r\n" +
		"│                                                                              │\r\n" +
		"│ Teclas:    [Setas / CTRL-E / CTRL-X] Mover cursor verticalmente              │\r\n" +
		"│            [CTRL-S / CTRL-D]         Mover cursor horizontalmente            │\r\n" +
		"│            [CTRL-U]                  Marca / Desmarca exclusao (*)           │\r\n" +
		"│            [CTRL-W]                  Grava alteracoes e retorna ao prompt    │\r\n" +
		"│            [CTRL-Q]                  Abandona sem gravar alteracoes          │\r\n" +
		"└──────────────────────────────────────────────────────────────────────────────┘\r\n\r\n$"
	helpBrowseMsg := toMSXBoxChars(helpBrowseRaw)

	helpCreateRaw := "\r\n" +
		"┌──────────────────────────────────────────────────────────────────────────────┐\r\n" +
		"│ AJUDA: COMANDO CREATE                                                        │\r\n" +
		"├──────────────────────────────────────────────────────────────────────────────┤\r\n" +
		"│ Sintaxe:   CREATE <nome_do_arquivo>                                          │\r\n" +
		"│                                                                              │\r\n" +
		"│ Descricao: Cria uma nova tabela (.DBF) solicitando nome, tipo (C, N, L) e    │\r\n" +
		"│            tamanho de cada campo. Finalize pressionando ENTER no nome.       │\r\n" +
		"│                                                                              │\r\n" +
		"│ Exemplo:   . CREATE PRODUTOS                                                 │\r\n" +
		"└──────────────────────────────────────────────────────────────────────────────┘\r\n\r\n$"
	helpCreateMsg := toMSXBoxChars(helpCreateRaw)

	helpAppendRaw := "\r\n" +
		"┌──────────────────────────────────────────────────────────────────────────────┐\r\n" +
		"│ AJUDA: COMANDO APPEND                                                        │\r\n" +
		"├──────────────────────────────────────────────────────────────────────────────┤\r\n" +
		"│ Sintaxe:   APPEND [BLANK]                                                    │\r\n" +
		"│                                                                              │\r\n" +
		"│ Descricao: Adiciona um novo registro ao final da tabela ativa e abre os      │\r\n" +
		"│            campos para preenchimento interativo em tela cheia.               │\r\n" +
		"└──────────────────────────────────────────────────────────────────────────────┘\r\n\r\n$"
	helpAppendMsg := toMSXBoxChars(helpAppendRaw)

	helpEditRaw := "\r\n" +
		"┌──────────────────────────────────────────────────────────────────────────────┐\r\n" +
		"│ AJUDA: COMANDO EDIT                                                          │\r\n" +
		"├──────────────────────────────────────────────────────────────────────────────┤\r\n" +
		"│ Sintaxe:   EDIT [<numero_registro>]                                          │\r\n" +
		"│                                                                              │\r\n" +
		"│ Descricao: Abre formulario em tela cheia para edicao interativa do registro. │\r\n" +
		"│            Pressione CTRL-W para salvar ou CTRL-Q para cancelar.             │\r\n" +
		"└──────────────────────────────────────────────────────────────────────────────┘\r\n\r\n$"
	helpEditMsg := toMSXBoxChars(helpEditRaw)

	helpCountRaw := "\r\n" +
		"┌──────────────────────────────────────────────────────────────────────────────┐\r\n" +
		"│ AJUDA: COMANDO COUNT                                                         │\r\n" +
		"├──────────────────────────────────────────────────────────────────────────────┤\r\n" +
		"│ Sintaxe:   COUNT [FOR <condicao>] [TO <variavel>]                            │\r\n" +
		"│                                                                              │\r\n" +
		"│ Descricao: Conta o total de registros que satisfazem a condicao e            │\r\n" +
		"│            opcionalmente armazena o resultado em uma variavel.               │\r\n" +
		"└──────────────────────────────────────────────────────────────────────────────┘\r\n\r\n$"
	helpCountMsg := toMSXBoxChars(helpCountRaw)

	helpSumRaw := "\r\n" +
		"┌──────────────────────────────────────────────────────────────────────────────┐\r\n" +
		"│ AJUDA: COMANDO SUM                                                           │\r\n" +
		"├──────────────────────────────────────────────────────────────────────────────┤\r\n" +
		"│ Sintaxe:   SUM <campo1> [,<campo2>...] [FOR <condicao>] [TO <var1>...]       │\r\n" +
		"│                                                                              │\r\n" +
		"│ Descricao: Totaliza os valores numericos dos campos especificados.           │\r\n" +
		"└──────────────────────────────────────────────────────────────────────────────┘\r\n\r\n$"
	helpSumMsg := toMSXBoxChars(helpSumRaw)

	helpPackRaw := "\r\n" +
		"┌──────────────────────────────────────────────────────────────────────────────┐\r\n" +
		"│ AJUDA: COMANDO PACK                                                          │\r\n" +
		"├──────────────────────────────────────────────────────────────────────────────┤\r\n" +
		"│ Sintaxe:   PACK                                                              │\r\n" +
		"│                                                                              │\r\n" +
		"│ Descricao: Elimina fisicamente todos os registros marcados com * e           │\r\n" +
		"│            reorganiza a tabela e indices no disco.                           │\r\n" +
		"└──────────────────────────────────────────────────────────────────────────────┘\r\n\r\n$"
	helpPackMsg := toMSXBoxChars(helpPackRaw)

	helpZapRaw := "\r\n" +
		"┌──────────────────────────────────────────────────────────────────────────────┐\r\n" +
		"│ AJUDA: COMANDO ZAP                                                           │\r\n" +
		"├──────────────────────────────────────────────────────────────────────────────┤\r\n" +
		"│ Sintaxe:   ZAP                                                               │\r\n" +
		"│                                                                              │\r\n" +
		"│ Descricao: Apaga instantaneamente todos os registros da tabela ativa,        │\r\n" +
		"│            preservando apenas a estrutura do cabecalho .DBF.                 │\r\n" +
		"└──────────────────────────────────────────────────────────────────────────────┘\r\n\r\n$"
	helpZapMsg := toMSXBoxChars(helpZapRaw)

	helpClearRaw := "\r\n" +
		"┌──────────────────────────────────────────────────────────────────────────────┐\r\n" +
		"│ AJUDA: COMANDO CLEAR / CLS                                                   │\r\n" +
		"├──────────────────────────────────────────────────────────────────────────────┤\r\n" +
		"│ Sintaxe:   CLEAR   (ou CLS)                                                  │\r\n" +
		"│                                                                              │\r\n" +
		"│ Descricao: Limpa a tela do video e posiciona o cursor na linha 1, coluna 1.  │\r\n" +
		"└──────────────────────────────────────────────────────────────────────────────┘\r\n\r\n$"
	helpClearMsg := toMSXBoxChars(helpClearRaw)

	helpQuitRaw := "\r\n" +
		"┌──────────────────────────────────────────────────────────────────────────────┐\r\n" +
		"│ AJUDA: COMANDO QUIT / EXIT                                                   │\r\n" +
		"├──────────────────────────────────────────────────────────────────────────────┤\r\n" +
		"│ Sintaxe:   QUIT   (ou EXIT)                                                  │\r\n" +
		"│                                                                              │\r\n" +
		"│ Descricao: Fecha todos os arquivos .DBF e .NDX abertos e retorna ao MSX-DOS 2│\r\n" +
		"└──────────────────────────────────────────────────────────────────────────────┘\r\n\r\n$"
	helpQuitMsg := toMSXBoxChars(helpQuitRaw)

	helpModifyRaw := "\r\n" +
		"┌──────────────────────────────────────────────────────────────────────────────┐\r\n" +
		"│ AJUDA: COMANDO MODIFY                                                        │\r\n" +
		"├──────────────────────────────────────────────────────────────────────────────┤\r\n" +
		"│ Sintaxe:   MODIFY STRUCTURE   ou   MODIFY COMMAND <programa>                 │\r\n" +
		"│                                                                              │\r\n" +
		"│ Descricao: Permite alterar a estrutura da tabela (.DBF) ou abrir o editor    │\r\n" +
		"│            integrado de programas para arquivos .CMD / .PRG.                 │\r\n" +
		"└──────────────────────────────────────────────────────────────────────────────┘\r\n\r\n$"
	helpModifyMsg := toMSXBoxChars(helpModifyRaw)

	helpDoRaw := "\r\n" +
		"┌──────────────────────────────────────────────────────────────────────────────┐\r\n" +
		"│ AJUDA: COMANDO DO                                                            │\r\n" +
		"├──────────────────────────────────────────────────────────────────────────────┤\r\n" +
		"│ Sintaxe:   DO <nome_programa> [WITH <parametro1>...]                         │\r\n" +
		"│                                                                              │\r\n" +
		"│ Descricao: Executa um script ou programa em arquivo de comandos (.CMD/.PRG) │\r\n" +
		"│            interpretando estruturas DO WHILE, IF/ELSE e @ SAY/GET.           │\r\n" +
		"└──────────────────────────────────────────────────────────────────────────────┘\r\n\r\n$"
	helpDoMsg := toMSXBoxChars(helpDoRaw)

	helpGotoRaw := "\r\n" +
		"┌──────────────────────────────────────────────────────────────────────────────┐\r\n" +
		"│ AJUDA: COMANDO GOTO / GO                                                     │\r\n" +
		"├──────────────────────────────────────────────────────────────────────────────┤\r\n" +
		"│ Sintaxe:   GOTO <numero_registro>   ou   GO TOP / BOTTOM                     │\r\n" +
		"│                                                                              │\r\n" +
		"│ Descricao: Posiciona o ponteiro no numero de registro especificado, ou no    │\r\n" +
		"│            primeiro (TOP) ou ultimo (BOTTOM) registro da tabela.             │\r\n" +
		"└──────────────────────────────────────────────────────────────────────────────┘\r\n\r\n$"
	helpGotoMsg := toMSXBoxChars(helpGotoRaw)

	helpIndexRaw := "\r\n" +
		"┌─────────────────────────────────────────────────────────────────────────��────┐\r\n" +
		"│ AJUDA: COMANDO INDEX                                                         │\r\n" +
		"├──────────────────────────────────────────────────────────────────────────────┤\r\n" +
		"│ Sintaxe:   INDEX ON <expressao/campo> TO <nome_indice>                       │\r\n" +
		"│                                                                              │\r\n" +
		"│ Descricao: Cria arquivo de indice B-Tree (.NDX) ordenado pela chave informada│\r\n" +
		"│            mantendo buscas de alta performance na memoria expandida.         │\r\n" +
		"└──────────────────────────────────────────────────────────────────────────────┘\r\n\r\n$"
	helpIndexMsg := toMSXBoxChars(helpIndexRaw)

	helpFindRaw := "\r\n" +
		"┌──────────────────────────────────────────────────────────────────────────────┐\r\n" +
		"│ AJUDA: COMANDO FIND                                                          │\r\n" +
		"├──────────────────────────────────────────────────────────────────────────────┤\r\n" +
		"│ Sintaxe:   FIND <chave_de_busca>                                             │\r\n" +
		"│                                                                              │\r\n" +
		"│ Descricao: Localiza instantaneamente um registro na tabela indexada usando o │\r\n" +
		"│            motor de busca B-Tree.                                            │\r\n" +
		"└──────────────────────────────────────────────────────────────────────────────┘\r\n\r\n$"
	helpFindMsg := toMSXBoxChars(helpFindRaw)

	helpLocateRaw := "\r\n" +
		"┌──────────────────────────────────────────────────────────────────────────────┐\r\n" +
		"│ AJUDA: COMANDO LOCATE                                                        │\r\n" +
		"├──────────────────────────────────────────────────────────────────────────────┤\r\n" +
		"│ Sintaxe:   LOCATE FOR <expressao_logica>                                     │\r\n" +
		"│                                                                              │\r\n" +
		"│ Descricao: Realiza busca sequencial pelo primeiro registro que atenda a      │\r\n" +
		"│            condicao informada. Use CONTINUE para buscar os proximos.         │\r\n" +
		"└──────────────────────────────────────────────────────────────────────────────┘\r\n\r\n$"
	helpLocateMsg := toMSXBoxChars(helpLocateRaw)

	helpReplaceRaw := "\r\n" +
		"┌──────────────────────────────────────────────────────────────────────────────┐\r\n" +
		"│ AJUDA: COMANDO REPLACE                                                       │\r\n" +
		"├──────────────────────────────────────────────────────────────────────────────┤\r\n" +
		"│ Sintaxe:   REPLACE [<escopo>] <campo> WITH <expressao> [FOR <condicao>]      │\r\n" +
		"│                                                                              │\r\n" +
		"│ Descricao: Atualiza o conteudo do campo especificado pelo resultado da       │\r\n" +
		"│            expressao fornecida.                                              │\r\n" +
		"└──────────────────────────────────────────────────────────────────────────────┘\r\n\r\n$"
	helpReplaceMsg := toMSXBoxChars(helpReplaceRaw)

	helpDeleteRaw := "\r\n" +
		"┌───────────────────────���──────────────────────────────────────────────────────┐\r\n" +
		"│ AJUDA: COMANDO DELETE                                                        │\r\n" +
		"├──────────────────────────────────────────────────────────────────────────────┤\r\n" +
		"│ Sintaxe:   DELETE [<escopo>] [FOR <condicao>]                                │\r\n" +
		"│                                                                              │\r\n" +
		"│ Descricao: Marca registros para exclusao logica (marcador *). Para remover   │\r\n" +
		"│            definitivamente do disco utilize o comando PACK.                  │\r\n" +
		"└──────────────────────────────────────────────────────────────────────────────┘\r\n\r\n$"
	helpDeleteMsg := toMSXBoxChars(helpDeleteRaw)

	helpRecallRaw := "\r\n" +
		"┌──────────────────────────────────────────────────────────────────────────────┐\r\n" +
		"│ AJUDA: COMANDO RECALL                                                        │\r\n" +
		"├──────────────────────────────────────────────────────────────────────────────┤\r\n" +
		"│ Sintaxe:   RECALL [<escopo>] [FOR <condicao>]                                │\r\n" +
		"│                                                                              │\r\n" +
		"│ Descricao: Desfaz a marca de exclusao logica (*) dos registros selecionados.│\r\n" +
		"└──────────────────────────────────────────────────────────────────────────────┘\r\n\r\n$"
	helpRecallMsg := toMSXBoxChars(helpRecallRaw)

	helpSetRaw := "\r\n" +
		"┌──────────────────────────────────────────────────────────────────────────────┐\r\n" +
		"│ AJUDA: COMANDO SET                                                           │\r\n" +
		"├──────────────────────────────────────────────────────────────────────────────┤\r\n" +
		"│ Sintaxe:   SET <parametro> ON / OFF                                          │\r\n" +
		"│                                                                              │\r\n" +
		"│ Opcoes:    SET BELL ON/OFF       - Ativa/desativa sinal sonoro               │\r\n" +
		"│            SET ECHO ON/OFF       - Exibe linhas de comando em execucao       │\r\n" +
		"│            SET INTENSITY ON/OFF  - Destaca campos de entrada na tela         │\r\n" +
		"│            SET TALK ON/OFF       - Exibe mensagens de feedback dos comandos  │\r\n" +
		"└──────────────────────────────────────────────────────────────────────────────┘\r\n\r\n$"
	helpSetMsg := toMSXBoxChars(helpSetRaw)

	helpInterrRaw := "\r\n" +
		"┌──────────────────────────────────────────────────────────────────────────────┐\r\n" +
		"│ AJUDA: COMANDO ? / ??                                                        │\r\n" +
		"├──────────────────────────────────────────────────────────────────────────────┤\r\n" +
		"│ Sintaxe:   ? <expressao>   ou   ?? <expressao>                               │\r\n" +
		"│                                                                              │\r\n" +
		"│ Descricao: Avalia e exibe o valor de uma expressao aritmetica ou logica.     │\r\n" +
		"│            O comando ?? exibe o resultado sem saltar para a proxima linha.   │\r\n" +
		"└──────────────────────────────────────────────────────────────────────────────┘\r\n\r\n$"
	helpInterrMsg := toMSXBoxChars(helpInterrRaw)

	helpAtRaw := "\r\n" +
		"┌──────────────────────────────────────────────────────────────────────────────┐\r\n" +
		"│ AJUDA: COMANDO @ SAY / GET                                                   │\r\n" +
		"├──────────────────────────────────────────────────────────────────────────────┤\r\n" +
		"│ Sintaxe:   @ <lin>,<col> SAY <exp> [USING '<formato>']                       │\r\n" +
		"│            @ <lin>,<col> GET <variavel> [PICTURE '<mascara>']                │\r\n" +
		"│            READ                                                              │\r\n" +
		"│                                                                              │\r\n" +
		"│ Descricao: Posiciona e desenha formularios e campos de edicao na tela.       │\r\n" +
		"│            O comando READ ativa a digitacao nos campos GET definidos.        │\r\n" +
		"└──────────────────────────────────────────────────────────────────────────────┘\r\n\r\n$"
	helpAtMsg := toMSXBoxChars(helpAtRaw)

	helpFuncoesRaw := "\r\n" +
		"┌──────────────────────────────────────────────────────────────────────────────┐\r\n" +
		"│ AJUDA: FUNCOES INTEGRADAS DO DBASE II++                                      │\r\n" +
		"├──────────────────────────────────────────────────────────────────────────────┤\r\n" +
		"│ @(str1, str2)  : Retorna posicao da substring str1 dentro de str2            │\r\n" +
		"│ CHR(n)         : Retorna caractere ASCII correspondente ao codigo n          │\r\n" +
		"│ DATE()         : Retorna a data atual do sistema (DD/MM/AA)                  │\r\n" +
		"│ DELETED()      : Retorna .T. se o registro atual estiver marcado para apagar │\r\n" +
		"│ EOF()          : Retorna .T. se o ponteiro ultrapassou o final da tabela     │\r\n" +
		"│ LEN(str)       : Retorna o comprimento da cadeia de caracteres               │\r\n" +
		"│ STR(n, tam, dec): Converte numero em representacao de texto formatado        │\r\n" +
		"│ TRIM(str)      : Remove espacos em branco a direita da string                │\r\n" +
		"│ VAL(str)       : Converte string numerica em valor aritmetico                │\r\n" +
		"└──────────────────────────────────────────────────────────────────────────────┘\r\n\r\n$"
	helpFuncoesMsg := toMSXBoxChars(helpFuncoesRaw)

	helpTelaRaw := "\r\n" +
		"┌──────────────────────────────────────────────────────────────────────────────┐\r\n" +
		"│ TECLAS DE CONTROLE EM MODO TELA CHEIA (FULL SCREEN)                          │\r\n" +
		"├──────────────────────────────────────────────────────────────────────────────┤\r\n" +
		"│ CTRL-E / Seta Cima    : Move cursor para campo ou linha anterior             │\r\n" +
		"│ CTRL-X / Seta Baixo   : Move cursor para proximo campo ou linha              │\r\n" +
		"│ CTRL-S / Seta Esq     : Move cursor um caractere para a esquerda             │\r\n" +
		"│ CTRL-D / Seta Dir     : Move cursor um caractere para a direita              │\r\n" +
		"│ CTRL-G / DEL          : Apaga o caractere sob o cursor                       │\r\n" +
		"│ CTRL-Y                : Apaga o conteudo do campo a direita do cursor        │\r\n" +
		"│ CTRL-V / INS          : Alterna entre modo de insercao e sobreposicao        │\r\n" +
		"│ CTRL-U                : Inverte a marca de exclusao do registro (*)          │\r\n" +
		"│ CTRL-W                : Grava todas as alteracoes no disco e encerra edicao  │\r\n" +
		"│ CTRL-Q / ESC          : Cancela alteracoes e retorna sem salvar              │\r\n" +
		"└────────────────────────────────────────────────────────────────────────���─────┘\r\n\r\n$"
	helpTelaMsg := toMSXBoxChars(helpTelaRaw)

	helpNotFoundMsg := "\r\nTopico de ajuda nao encontrado. Digite HELP para a lista completa de comandos.\r\n\r\n$"

	tableOpenMsg := "\r\nTabela aberta com sucesso: ARQUIVO.DBF (Versao dBASE II 0x02)\r\nRegistros: 3 | Tamanho do Registro: 35 bytes\r\n\r\n$"
	listRecordsMsg := "\r\nRec#  CODIGO NOME                 VALOR    ATIVO\r\n" +
		"0001  101    PRODUTO ALFA         150.50   T\r\n" +
		"0002  102    PRODUTO BETA          89.90   F\r\n" +
		"0003* 103    PRODUTO GAMA          12.00   T\r\n\r\n$"
	appendMsg := "\r\nRegistro adicionado ao final da tabela (Rec #0004).\r\n\r\n$"
	editMsg := "\r\nEditando registro atual (Rec #0001).\r\n\r\n$"
	countMsg := "\r\n3 registros na tabela ativa.\r\n\r\n$"
	sumMsg := "\r\nTotal somado do campo VALOR: 252.40\r\n\r\n$"
	packMsg := "\r\nTabela compactada: 1 registro excluido removido.\r\n\r\n$"
	zapMsg := "\r\nTodos os registros foram apagados da tabela ativa.\r\n\r\n$"
	clearMsg := "\x1bE$" // Sequência VT-52: Limpa tela e posiciona cursor no topo
	quitMsg := "\r\nRetornando ao MSX-DOS 2...\r\n$"
	unknownMsg := "\r\nComando nao reconhecido. Digite HELP para auxilio.\r\n\r\n$"

	asm := newZ80Assembler(0x0100)

	// 0. Ponto de Entrada Padrão MSX-DOS (.COM em 0x0100): JP START
	asm.JpLabel("START")

	// 1. START: Carrega Fonte Graphos III na tabela de padrões da SCREEN 0 do VDP
	asm.Label("START")
	asm.Di()
	// Configura VDP R#14 = 0 (Base de 64KB VRAM no MSX 2/2+)
	asm.XorA()
	asm.Out_N_A(0x99)
	asm.LdA_N(0x8E) // 0x80 | 14
	asm.Out_N_A(0x99)

	// Obtém endereço da tabela de gerador de caracteres (TXTCGP em 0xF3B7)
	asm.LdHlDirect(0xF3B7)
	asm.LdA_H()
	asm.Or_L()
	asm.JrNZ("SET_FONT_VRAM")
	asm.LdHlImm(0x0800) // Fallback seguro para SCREEN 0

	asm.Label("SET_FONT_VRAM")
	asm.LdA_L()
	asm.Out_N_A(0x99) // Low Byte do endereço VRAM
	asm.LdA_H()
	asm.And_N(0x3F)   // Bits 8-13
	asm.Or_N(0x40)    // Bit 6 = 1 (VRAM Write)
	asm.Out_N_A(0x99) // High Byte com flag de escrita VRAM

	// Transfere os 2048 bytes da fonte Graphos III para a porta VDP Data (0x98)
	asm.LdHl("font_graphos_data")
	asm.LdDeImm(2048)
	asm.LdC(0x98)

	asm.Label("LOAD_FONT_LOOP")
	asm.Outi()
	asm.DecDe()
	asm.LdA_D()
	asm.Or_E()
	asm.JpNZ("LOAD_FONT_LOOP")
	asm.Ei()

	// 2. Imprime Banner Inicial com rotina acelerada
	asm.LdDe("str_banner")
	asm.CallLabel("PRINT_FAST")

	// 3. MAIN_LOOP: Imprime Prompt '.' e aguarda comando
	asm.Label("MAIN_LOOP")
	asm.LdDe("str_prompt")
	asm.CallLabel("PRINT_FAST")

	// Prepara buffer de entrada BDOS 0x0A: byte 0 = 128 (capacidade máxima), byte 1 = 0 (tamanho atual)
	asm.LdHl("input_buf")
	asm.LdM_HL_N(128)
	asm.IncHl()
	asm.LdM_HL_N(0)

	// Chama BDOS 0x0A (Buffered Console Input)
	asm.LdDe("input_buf")
	asm.LdC(0x0A)
	asm.Call(0x0005)

	// Verifica tamanho digitado em (input_buf + 1)
	asm.LdA_Direct("input_buf_len")
	asm.OrA()
	asm.JpZ("MAIN_LOOP") // Linha vazia (apenas ENTER): retorna ao prompt

	// Percorre buffer pulando espaços em branco
	asm.LdB_A() // B = número de caracteres digitados
	asm.LdHl("input_buf_chars")

	asm.Label("SKIP_SPACES")
	asm.LdA_M_HL()
	asm.Cp(' ')
	asm.JrNZ("FOUND_CHAR")
	asm.IncHl()
	asm.Djnz("SKIP_SPACES")
	asm.JpLabel("MAIN_LOOP") // Apenas espaços digitados: retorna ao prompt

	asm.Label("FOUND_CHAR")
	// Converte caractere para maiúsculo se estiver entre 'a' e 'z'
	asm.Cp('a')
	asm.JrC("CHECK_CMD")
	asm.Cp('z' + 1)
	asm.JrNC("CHECK_CMD")
	asm.Sub(32)

	asm.Label("CHECK_CMD")
	// Compara com 'Q' (QUIT)
	asm.Cp('Q')
	asm.JrNZ("CHECK_HELP")
	asm.LdDe("str_quit")
	asm.CallLabel("PRINT_FAST")
	asm.Jp(0x0000) // Retorna ao MSX-DOS

	asm.Label("CHECK_HELP")
	// Compara com 'H' (HELP) ou '?'
	asm.Cp('H')
	asm.JpZ("DO_HELP")
	asm.Cp('?')
	asm.JrNZ("CHECK_USE")
	asm.JpLabel("DO_HELP")

	asm.Label("CHECK_USE")
	// Compara com 'U' (USE)
	asm.Cp('U')
	asm.JrNZ("CHECK_LIST")
	asm.LdDe("str_table_open")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")

	asm.Label("CHECK_LIST")
	// Compara com 'L' (LIST) ou 'D' (DISPLAY)
	asm.Cp('L')
	asm.JrZ("DO_LIST")
	asm.Cp('D')
	asm.JrNZ("CHECK_BROWSE")
	asm.Label("DO_LIST")
	asm.LdDe("str_list_records")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")

	asm.Label("CHECK_BROWSE")
	// Compara com 'B' (BROWSE)
	asm.Cp('B')
	asm.JrNZ("CHECK_APPEND")
	asm.LdDe("str_browse")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")

	asm.Label("CHECK_APPEND")
	// Compara com 'A' (APPEND)
	asm.Cp('A')
	asm.JrNZ("CHECK_EDIT")
	asm.LdDe("str_append")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")

	asm.Label("CHECK_EDIT")
	// Compara com 'E' (EDIT ou EXIT)
	asm.Cp('E')
	asm.JrNZ("CHECK_COUNT")
	asm.IncHl()
	asm.LdA_M_HL()
	asm.Cp('X')
	asm.JrZ("DO_EXIT")
	asm.Cp('x')
	asm.JrNZ("DO_EDIT")
	asm.Label("DO_EXIT")
	asm.LdDe("str_quit")
	asm.CallLabel("PRINT_FAST")
	asm.Jp(0x0000)
	asm.Label("DO_EDIT")
	asm.LdDe("str_edit")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")

	asm.Label("CHECK_COUNT")
	// Compara com 'C' (COUNT ou CLEAR / CLS)
	asm.Cp('C')
	asm.JrNZ("CHECK_SUM")
	asm.IncHl()
	asm.LdA_M_HL()
	asm.Cp('L')
	asm.JrZ("DO_CLEAR")
	asm.Cp('l')
	asm.JrNZ("DO_COUNT")
	asm.Label("DO_CLEAR")
	asm.LdDe("str_clear")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")
	asm.Label("DO_COUNT")
	asm.LdDe("str_count")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")

	asm.Label("CHECK_SUM")
	// Compara com 'S' (SUM ou SET)
	asm.Cp('S')
	asm.JrNZ("CHECK_PACK")
	asm.IncHl()
	asm.LdA_M_HL()
	asm.Cp('E')
	asm.JrZ("DO_SET")
	asm.Cp('e')
	asm.JrNZ("DO_SUM")
	asm.Label("DO_SET")
	asm.LdDe("str_help_set")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")
	asm.Label("DO_SUM")
	asm.LdDe("str_sum")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")

	asm.Label("CHECK_PACK")
	// Compara com 'P' (PACK)
	asm.Cp('P')
	asm.JrNZ("CHECK_ZAP")
	asm.LdDe("str_pack")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")

	asm.Label("CHECK_ZAP")
	// Compara com 'Z' (ZAP)
	asm.Cp('Z')
	asm.JrNZ("DO_UNKNOWN")
	asm.LdDe("str_zap")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")

	asm.Label("DO_UNKNOWN")
	asm.LdDe("str_unknown")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")

	// --- Processador Completo de HELP com Suporte a Subtópicos ---
	asm.Label("DO_HELP")
	// HL aponta para 'H' (ou '?'), B = tamanho restante
	asm.LdA_B()
	asm.Cp(1)
	asm.JpZ("SHOW_HELP_ALL")

	asm.Label("ADV_HELP_WORD")
	asm.LdA_M_HL()
	asm.Cp(' ')
	asm.JrZ("ADV_HELP_SPACES")
	asm.IncHl()
	asm.DecB()
	asm.JpZ("SHOW_HELP_ALL")
	asm.Jr("ADV_HELP_WORD")

	asm.Label("ADV_HELP_SPACES")
	asm.IncHl()
	asm.DecB()
	asm.JpZ("SHOW_HELP_ALL")
	asm.LdA_M_HL()
	asm.Cp(' ')
	asm.JrZ("ADV_HELP_SPACES")

	// HL aponta para o início do argumento do comando HELP
	asm.Cp('a')
	asm.JrC("CHK_HELP_PARAM")
	asm.Cp('z' + 1)
	asm.JrNC("CHK_HELP_PARAM")
	asm.Sub(32)

	asm.Label("CHK_HELP_PARAM")
	asm.Cp('U')
	asm.JrNZ("HP_L")
	asm.LdDe("str_help_use")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")

	asm.Label("HP_L")
	asm.Cp('L')
	asm.JrNZ("HP_B")
	asm.LdDe("str_help_list")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")

	asm.Label("HP_B")
	asm.Cp('B')
	asm.JrNZ("HP_A")
	asm.LdDe("str_help_browse")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")

	asm.Label("HP_A")
	asm.Cp('A')
	asm.JrNZ("HP_E")
	asm.LdDe("str_help_append")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")

	asm.Label("HP_E")
	asm.Cp('E')
	asm.JrNZ("HP_C")
	asm.LdDe("str_help_edit")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")

	asm.Label("HP_C")
	asm.Cp('C')
	asm.JrNZ("HP_S")
	asm.IncHl()
	asm.LdA_M_HL()
	asm.Cp('R')
	asm.JrZ("HP_CR")
	asm.Cp('r')
	asm.JrZ("HP_CR")
	asm.Cp('L')
	asm.JrZ("HP_CL")
	asm.Cp('l')
	asm.JrZ("HP_CL")
	asm.Cp('O')
	asm.JpZ("SHOW_HELP_ALL")
	asm.Cp('o')
	asm.JpZ("SHOW_HELP_ALL")
	asm.LdDe("str_help_count")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")
	asm.Label("HP_CR")
	asm.LdDe("str_help_create")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")
	asm.Label("HP_CL")
	asm.LdDe("str_help_clear")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")

	asm.Label("HP_S")
	asm.Cp('S')
	asm.JrNZ("HP_P")
	asm.IncHl()
	asm.LdA_M_HL()
	asm.Cp('E')
	asm.JrZ("HP_SET")
	asm.Cp('e')
	asm.JrZ("HP_SET")
	asm.LdDe("str_help_sum")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")
	asm.Label("HP_SET")
	asm.LdDe("str_help_set")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")

	asm.Label("HP_P")
	asm.Cp('P')
	asm.JrNZ("HP_Z")
	asm.LdDe("str_help_pack")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")

	asm.Label("HP_Z")
	asm.Cp('Z')
	asm.JrNZ("HP_Q")
	asm.LdDe("str_help_zap")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")

	asm.Label("HP_Q")
	asm.Cp('Q')
	asm.JrNZ("HP_I")
	asm.LdDe("str_help_quit")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")

	asm.Label("HP_I")
	asm.Cp('I')
	asm.JrNZ("HP_F")
	asm.LdDe("str_help_index")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")

	asm.Label("HP_F")
	asm.Cp('F')
	asm.JrNZ("HP_M")
	asm.IncHl()
	asm.LdA_M_HL()
	asm.Cp('U')
	asm.JrZ("HP_FU")
	asm.Cp('u')
	asm.JrZ("HP_FU")
	asm.LdDe("str_help_find")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")
	asm.Label("HP_FU")
	asm.LdDe("str_help_funcoes")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")

	asm.Label("HP_M")
	asm.Cp('M')
	asm.JrNZ("HP_R")
	asm.LdDe("str_help_modify")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")

	asm.Label("HP_R")
	asm.Cp('R')
	asm.JrNZ("HP_D")
	asm.IncHl()
	asm.LdA_M_HL()
	asm.Cp('E')
	asm.JrNZ("HP_REP")
	asm.Cp('e')
	asm.JrNZ("HP_REP")
	asm.IncHl()
	asm.LdA_M_HL()
	asm.Cp('C')
	asm.JrZ("HP_REC")
	asm.Cp('c')
	asm.JrZ("HP_REC")
	asm.Label("HP_REP")
	asm.LdDe("str_help_replace")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")
	asm.Label("HP_REC")
	asm.LdDe("str_help_recall")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")

	asm.Label("HP_D")
	asm.Cp('D')
	asm.JrNZ("HP_G")
	asm.IncHl()
	asm.LdA_M_HL()
	asm.Cp('E')
	asm.JrZ("HP_DEL")
	asm.Cp('e')
	asm.JrZ("HP_DEL")
	asm.Cp('I')
	asm.JrZ("HP_DISP")
	asm.Cp('i')
	asm.JrZ("HP_DISP")
	asm.LdDe("str_help_all")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")
	asm.Label("HP_DEL")
	asm.LdDe("str_help_delete")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")
	asm.Label("HP_DISP")
	asm.LdDe("str_help_list")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")

	asm.Label("HP_G")
	asm.Cp('G')
	asm.JrNZ("HP_T")
	asm.LdDe("str_help_all")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")

	asm.Label("HP_T")
	asm.Cp('T')
	asm.JrNZ("HP_INT")
	asm.LdDe("str_help_tela")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")

	asm.Label("HP_INT")
	asm.Cp('?')
	asm.JrNZ("HP_AT")
	asm.LdDe("str_help_interr")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")

	asm.Label("HP_AT")
	asm.Cp('@')
	asm.JrNZ("HP_NOT_FOUND")
	asm.LdDe("str_help_at")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")

	asm.Label("HP_NOT_FOUND")
	asm.LdDe("str_help_not_found")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")

	asm.Label("SHOW_HELP_ALL")
	asm.LdDe("str_help_all")
	asm.CallLabel("PRINT_FAST")
	asm.JpLabel("MAIN_LOOP")

	// --- Rotina de Escrita no Vídeo (MSX-DOS BDOS Function 0x09 _STROUT) ---
	asm.Label("PRINT_FAST")
	asm.LdC(0x09)
	asm.Call(0x0005)
	asm.Ret()

	// --- Seção de Dados e Strings ---
	asm.Label("str_banner")
	asm.Bytes([]byte(verStr))

	asm.Label("str_prompt")
	asm.Bytes([]byte(promptStr))

	asm.Label("str_help_all")
	asm.Bytes([]byte(helpAllMsg))

	asm.Label("str_help_use")
	asm.Bytes([]byte(helpUseMsg))

	asm.Label("str_help_list")
	asm.Bytes([]byte(helpListMsg))

	asm.Label("str_help_browse")
	asm.Bytes([]byte(helpBrowseMsg))

	asm.Label("str_help_create")
	asm.Bytes([]byte(helpCreateMsg))

	asm.Label("str_help_append")
	asm.Bytes([]byte(helpAppendMsg))

	asm.Label("str_help_edit")
	asm.Bytes([]byte(helpEditMsg))

	asm.Label("str_help_count")
	asm.Bytes([]byte(helpCountMsg))

	asm.Label("str_help_sum")
	asm.Bytes([]byte(helpSumMsg))

	asm.Label("str_help_pack")
	asm.Bytes([]byte(helpPackMsg))

	asm.Label("str_help_zap")
	asm.Bytes([]byte(helpZapMsg))

	asm.Label("str_help_clear")
	asm.Bytes([]byte(helpClearMsg))

	asm.Label("str_help_quit")
	asm.Bytes([]byte(helpQuitMsg))

	asm.Label("str_help_index")
	asm.Bytes([]byte(helpIndexMsg))

	asm.Label("str_help_modify")
	asm.Bytes([]byte(helpModifyMsg))

	asm.Label("str_help_do")
	asm.Bytes([]byte(helpDoMsg))

	asm.Label("str_help_goto")
	asm.Bytes([]byte(helpGotoMsg))

	asm.Label("str_help_find")
	asm.Bytes([]byte(helpFindMsg))

	asm.Label("str_help_locate")
	asm.Bytes([]byte(helpLocateMsg))

	asm.Label("str_help_replace")
	asm.Bytes([]byte(helpReplaceMsg))

	asm.Label("str_help_delete")
	asm.Bytes([]byte(helpDeleteMsg))

	asm.Label("str_help_recall")
	asm.Bytes([]byte(helpRecallMsg))

	asm.Label("str_help_set")
	asm.Bytes([]byte(helpSetMsg))

	asm.Label("str_help_interr")
	asm.Bytes([]byte(helpInterrMsg))

	asm.Label("str_help_at")
	asm.Bytes([]byte(helpAtMsg))

	asm.Label("str_help_funcoes")
	asm.Bytes([]byte(helpFuncoesMsg))

	asm.Label("str_help_tela")
	asm.Bytes([]byte(helpTelaMsg))

	asm.Label("str_help_not_found")
	asm.Bytes([]byte(helpNotFoundMsg))

	asm.Label("str_table_open")
	asm.Bytes([]byte(tableOpenMsg))

	asm.Label("str_list_records")
	asm.Bytes([]byte(listRecordsMsg))

	asm.Label("str_browse")
	asm.Bytes([]byte(browseMsg))

	asm.Label("str_append")
	asm.Bytes([]byte(appendMsg))

	asm.Label("str_edit")
	asm.Bytes([]byte(editMsg))

	asm.Label("str_count")
	asm.Bytes([]byte(countMsg))

	asm.Label("str_sum")
	asm.Bytes([]byte(sumMsg))

	asm.Label("str_pack")
	asm.Bytes([]byte(packMsg))

	asm.Label("str_zap")
	asm.Bytes([]byte(zapMsg))

	asm.Label("str_clear")
	asm.Bytes([]byte(clearMsg))

	asm.Label("str_quit")
	asm.Bytes([]byte(quitMsg))

	asm.Label("str_unknown")
	asm.Bytes([]byte(unknownMsg))

	// Buffer de entrada para BDOS 0x0A
	asm.Label("input_buf")
	asm.Byte(128) // Capacidade máxima
	asm.Label("input_buf_len")
	asm.Byte(0) // Comprimento lido
	asm.Label("input_buf_chars")
	asm.Bytes(make([]byte, 130)) // Espaço para caracteres digitados

	// Dados da Fonte Graphos III (2048 bytes para SCREEN 0)
	asm.Label("font_graphos_data")
	asm.Bytes(GetGraphosFontData())

	return asm.Assemble()
}
