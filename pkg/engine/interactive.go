// Package engine implementa o motor de execução de comandos dBASE II.
package engine

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"dbase2a/pkg/dbf"
)

// ReadLine exibe um prompt (se fornecido) e lê uma linha da entrada configurada no motor.
func (e *Engine) ReadLine(prompt string) (string, error) {
	if prompt != "" {
		e.Printf("%s", prompt)
	}
	if e.In == nil {
		e.In = os.Stdin
	}
	if e.reader == nil {
		e.reader = bufio.NewReader(e.In)
	}

	line, err := e.reader.ReadString('\n')
	if err != nil {
		if len(line) > 0 {
			return strings.TrimRight(line, "\r\n"), nil
		}
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

// SetIn atualiza a fonte de entrada de dados do motor (útil para testes automatizados).
func (e *Engine) SetIn(in io.Reader) {
	e.In = in
	if in != nil {
		e.reader = bufio.NewReader(in)
	} else {
		e.reader = nil
	}
}

// ParseFieldDefinition analisa uma linha de definição de campo no formato dBASE II.
// Formatos aceitos:
//   - "NOME, C, 30"
//   - "NOME,C,30"
//   - "VALOR, N, 10, 2"
//   - "VALOR,N,10,2"
//   - "ATIVO, L" ou "ATIVO, L, 1"
//   - "NOME C 30" (separado por espaços)
//   - "VALOR N 10 2"
func ParseFieldDefinition(line string) (dbf.Field, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return dbf.Field{}, fmt.Errorf("linha de campo vazia")
	}

	var parts []string
	if strings.Contains(line, ",") {
		rawParts := strings.Split(line, ",")
		for _, p := range rawParts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				parts = append(parts, trimmed)
			}
		}
	} else {
		parts = strings.Fields(line)
	}

	if len(parts) < 2 {
		return dbf.Field{}, fmt.Errorf("definição incompleta. Utilize: NOME, TIPO, TAMANHO [, DECIMAIS]")
	}

	// 1. Nome do campo
	name := strings.ToUpper(strings.Trim(parts[0], "\"'`"))
	if len(name) == 0 {
		return dbf.Field{}, fmt.Errorf("nome do campo não pode ser vazio")
	}
	if len(name) > 10 {
		return dbf.Field{}, fmt.Errorf("nome de campo '%s' ultrapassa o limite de 10 caracteres", name)
	}

	// Valida caracteres válidos no nome
	for i, r := range name {
		if i == 0 && (r < 'A' || r > 'Z') {
			return dbf.Field{}, fmt.Errorf("nome de campo '%s' deve iniciar com letra", name)
		}
		if (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '_' {
			return dbf.Field{}, fmt.Errorf("nome de campo '%s' contém caractere inválido '%c'", name, r)
		}
	}

	// 2. Tipo do campo
	typeStr := strings.ToUpper(parts[1])
	var fType dbf.FieldType
	switch typeStr {
	case "C", "CHAR", "CHARACTER", "CARACTERE":
		fType = dbf.TypeCharacter
	case "N", "NUM", "NUMERIC", "NUMÉRIC", "NUMERICO", "NUMÉRICO":
		fType = dbf.TypeNumeric
	case "L", "LOG", "LOGICAL", "LÓGIC", "LOGICO", "LÓGICO":
		fType = dbf.TypeLogical
	default:
		return dbf.Field{}, fmt.Errorf("tipo de campo inválido '%s'. Deve ser C (Caracter), N (Numérico) ou L (Lógico)", typeStr)
	}

	// 3. Tamanho (Length)
	var length uint8 = 1
	if fType == dbf.TypeLogical {
		length = 1
	} else {
		if len(parts) < 3 {
			return dbf.Field{}, fmt.Errorf("tamanho obrigatório para campo tipo '%c'", fType)
		}
		l, err := strconv.Atoi(parts[2])
		if err != nil || l < 1 || l > 254 {
			return dbf.Field{}, fmt.Errorf("tamanho inválido '%s'. Deve ser um número entre 1 e 254", parts[2])
		}
		length = uint8(l)
	}

	// 4. Casas Decimais (Decimals)
	var decimals uint8 = 0
	if fType == dbf.TypeNumeric {
		if len(parts) >= 4 {
			d, err := strconv.Atoi(parts[3])
			if err != nil || d < 0 || d > 19 {
				return dbf.Field{}, fmt.Errorf("casas decimais inválidas '%s'. Deve ser entre 0 e 19", parts[3])
			}
			if d >= int(length)-1 && length > 1 {
				return dbf.Field{}, fmt.Errorf("casas decimais (%d) devem ser menores que o tamanho total (%d)", d, length)
			}
			decimals = uint8(d)
		}
	}

	field := dbf.Field{
		Name:     name,
		Type:     fType,
		Length:   length,
		Decimals: decimals,
	}

	if err := field.Validate(); err != nil {
		return dbf.Field{}, err
	}

	return field, nil
}

// FormatFieldValue formata e ajusta um valor textual de acordo com as regras do tipo do campo.
func FormatFieldValue(val string, f dbf.Field) string {
	switch f.Type {
	case dbf.TypeCharacter:
		if len(val) > int(f.Length) {
			return val[:f.Length]
		}
		return val
	case dbf.TypeNumeric:
		val = strings.TrimSpace(val)
		if val == "" {
			return "0"
		}
		if f.Decimals > 0 {
			if num, err := strconv.ParseFloat(val, 64); err == nil {
				return fmt.Sprintf(fmt.Sprintf("%%%d.%df", f.Length, f.Decimals), num)
			}
		}
		return val
	case dbf.TypeLogical:
		upper := strings.ToUpper(strings.TrimSpace(val))
		if upper == "T" || upper == "Y" || upper == "S" || upper == "SIM" || upper == "1" || upper == ".T." || upper == ".Y." {
			return "T"
		} else if upper == "F" || upper == "N" || upper == "NAO" || upper == "NÃO" || upper == "0" || upper == ".F." || upper == ".N." {
			return "F"
		}
		return "F"
	}
	return val
}

// InteractiveCreate executa o fluxo passo a passo de criação de tabela dBASE II com cabeçalho de estrutura e prompt de inclusão.
func (e *Engine) InteractiveCreate(filePath string) error {
	if strings.TrimSpace(filePath) == "" {
		line, err := e.ReadLine("ENTER FILENAME: ")
		if err != nil || strings.TrimSpace(line) == "" {
			e.Println("Criação de arquivo cancelada.")
			return nil
		}
		filePath = strings.TrimSpace(line)
	}

	if !strings.HasSuffix(strings.ToUpper(filePath), ".DBF") {
		filePath += ".DBF"
	}

	e.Println("ENTER RECORD STRUCTURE AS FOLLOWS:")
	e.Println("FIELD  NAME,TYPE,WIDTH,DECIMAL PLACES")

	var fields []dbf.Field
	fieldNum := 1

	for fieldNum <= dbf.MaxFields {
		prompt := fmt.Sprintf("%03d    ", fieldNum)
		line, err := e.ReadLine(prompt)
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if line == "" {
			if len(fields) == 0 {
				e.Println("Criação de arquivo cancelada (nenhum campo informado).")
				return nil
			}
			// Fim da definição dos campos
			break
		}

		field, err := ParseFieldDefinition(line)
		if err != nil {
			e.Printf("Erro no campo: %v\n", err)
			e.Println("Exemplo válido: NOME,C,30 ou SALARIO,N,10,2 ou ATIVO,L")
			continue
		}

		// Checa duplicidade
		duplicate := false
		for _, existing := range fields {
			if strings.EqualFold(existing.Name, field.Name) {
				e.Printf("Campo '%s' já foi definido nesta tabela.\n", field.Name)
				duplicate = true
				break
			}
		}
		if duplicate {
			continue
		}

		fields = append(fields, field)
		fieldNum++
	}

	if len(fields) == 0 {
		e.Println("Nenhum campo definido. Criação cancelada.")
		return nil
	}

	tbl, err := dbf.CreateTable(filePath, fields)
	if err != nil {
		return fmt.Errorf("erro ao criar tabela '%s': %w", filePath, err)
	}

	if e.ActiveTable != nil {
		_ = e.ActiveTable.Close()
	}

	e.ActiveTable = tbl
	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	e.ActiveAlias = strings.ToUpper(strings.TrimSuffix(base, ext))
	e.CurrentRecordIndex = 1

	// Pergunta clássica do dBASE II se deseja entrar com dados agora
	ans, err := e.ReadLine("INPUT DATA NOW? (Y/N) ")
	if err != nil {
		return nil
	}
	ans = strings.TrimSpace(strings.ToUpper(ans))
	if ans == "Y" || ans == "YES" || ans == "S" || ans == "SIM" {
		return e.InteractiveAppend()
	}

	return nil
}

// InteractiveAppend executa a tela simples e contínua de inclusão de novos registros.
func (e *Engine) InteractiveAppend() error {
	if e.ActiveTable == nil {
		return fmt.Errorf("nenhuma tabela aberta para APPEND")
	}

	countAdded := 0
	for {
		recNo := len(e.ActiveTable.Records) + 1
		e.Printf("RECORD # %05d\n", recNo)

		values := make(map[string]string)
		aborted := false

		for i, f := range e.ActiveTable.Header.Fields {
			fieldPrompt := fmt.Sprintf("%-10s : ", f.Name)
			val, err := e.ReadLine(fieldPrompt)
			if err != nil {
				aborted = true
				break
			}

			// Se for o primeiro campo do registro e o usuário pressionar Enter vazio, encerra o APPEND
			if i == 0 && strings.TrimSpace(val) == "" {
				aborted = true
				break
			}

			values[f.Name] = FormatFieldValue(val, f)
		}

		if aborted {
			break
		}

		rec := dbf.NewRecord(e.ActiveTable.Header.Fields)
		for k, v := range values {
			rec.Set(k, v)
		}
		e.ActiveTable.Records = append(e.ActiveTable.Records, rec)
		e.ActiveTable.Header.RecordCount = uint16(len(e.ActiveTable.Records))
		e.CurrentRecordIndex = len(e.ActiveTable.Records)
		countAdded++
	}

	if countAdded > 0 {
		if err := e.ActiveTable.Save(); err != nil {
			return fmt.Errorf("falha ao salvar registros no disco: %w", err)
		}
	}

	if e.Config.Talk {
		e.Printf("%05d REGISTRO(S) ADICIONADO(S)\n", countAdded)
	}
	return nil
}

// InteractiveEdit permite a edição interativa campo a campo de um registro existente.
func (e *Engine) InteractiveEdit(recNo int) error {
	if e.ActiveTable == nil {
		return fmt.Errorf("nenhuma tabela aberta para EDIT")
	}

	if recNo <= 0 {
		recNo = e.CurrentRecordIndex
	}
	if recNo < 1 || recNo > len(e.ActiveTable.Records) {
		return fmt.Errorf("registro %d fora dos limites (1 a %d)", recNo, len(e.ActiveTable.Records))
	}

	e.CurrentRecordIndex = recNo
	rec := e.ActiveTable.Records[recNo-1]

	e.Printf("RECORD # %05d\n", recNo)
	for _, f := range e.ActiveTable.Header.Fields {
		curVal := rec.Get(f.Name)
		fieldPrompt := fmt.Sprintf("%-10s [%s] : ", f.Name, curVal)
		val, err := e.ReadLine(fieldPrompt)
		if err != nil {
			break
		}
		val = strings.TrimRight(val, "\r\n")
		if val != "" {
			rec.Set(f.Name, FormatFieldValue(val, f))
		}
	}

	if err := e.ActiveTable.Save(); err != nil {
		return fmt.Errorf("erro ao salvar edição: %w", err)
	}

	if e.Config.Talk {
		e.Printf("Registro %d atualizado.\n", recNo)
	}
	return nil
}
