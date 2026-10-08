package engine

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"dbase2a/pkg/dbf"
	"dbase2a/pkg/parser"
)

func (e *Engine) execStatement(stmt *parser.Statement) error {
	if stmt == nil {
		return nil
	}

	switch stmt.Type {
	case parser.CmdUse:
		return e.cmdUse(stmt)
	case parser.CmdCreate:
		return e.cmdCreate(stmt)
	case parser.CmdList:
		return e.cmdList(stmt)
	case parser.CmdDisplay:
		return e.cmdDisplay(stmt)
	case parser.CmdAppend:
		return e.cmdAppend(stmt)
	case parser.CmdEdit:
		return e.cmdEdit(stmt)
	case parser.CmdDelete:
		return e.cmdDelete(stmt)
	case parser.CmdRecall:
		return e.cmdRecall(stmt)
	case parser.CmdPack:
		return e.cmdPack(stmt)
	case parser.CmdZap:
		return e.cmdZap(stmt)
	case parser.CmdReplace:
		return e.cmdReplace(stmt)
	case parser.CmdLocate:
		return e.cmdLocate(stmt)
	case parser.CmdContinue:
		return e.cmdContinue(stmt)
	case parser.CmdCount:
		return e.cmdCount(stmt)
	case parser.CmdSum:
		return e.cmdSum(stmt)
	case parser.CmdDo:
		return e.cmdDo(stmt)
	case parser.CmdSet:
		return e.cmdSet(stmt)
	case parser.CmdQuit:
		return e.cmdQuit(stmt)
	case parser.CmdHelp:
		return e.cmdHelp(stmt)
	case parser.CmdClear:
		e.Printf("\033[H\033[2J") // Limpeza ANSI
		return nil
	default:
		return fmt.Errorf("comando não reconhecido ou ainda não implementado: %s", stmt.Raw)
	}
}

func (e *Engine) cmdUse(stmt *parser.Statement) error {
	if stmt.Target == "" {
		// USE sem parâmetros fecha a tabela atual
		if e.ActiveTable != nil {
			_ = e.ActiveTable.Save()
			e.ActiveTable = nil
			e.ActiveAlias = ""
			if e.Config.Talk {
				e.Println("Nenhum arquivo em uso.")
			}
		}
		return nil
	}

	filePath := stmt.Target
	if filepath.Ext(filePath) == "" {
		filePath += ".DBF"
	}

	tbl, err := dbf.OpenTable(filePath)
	if err != nil {
		return err
	}

	if e.ActiveTable != nil {
		_ = e.ActiveTable.Save()
	}

	e.ActiveTable = tbl
	e.ActiveAlias = strings.ToUpper(filepath.Base(filePath))

	if e.Storage != nil {
		_ = e.Storage.CatalogTable(e.ActiveAlias, filePath, len(tbl.Records), len(tbl.Header.Fields), int(tbl.Header.RecordSize), tbl.Header.LastUpdate)
	}

	if e.Config.Talk {
		e.Printf("Tabela %s aberta com sucesso (%d registros, %d campos).\n", e.ActiveAlias, len(tbl.Records), len(tbl.Header.Fields))
	}
	return nil
}

func (e *Engine) cmdCreate(stmt *parser.Statement) error {
	if stmt.Target == "" {
		return fmt.Errorf("uso: CREATE <nome_arquivo>")
	}
	filePath := stmt.Target
	if filepath.Ext(filePath) == "" {
		filePath += ".DBF"
	}

	// Criação de estrutura padrão básica caso chamada programaticamente
	defaultFields := []dbf.Field{
		{Name: "ID", Type: dbf.TypeNumeric, Length: 6, Decimals: 0},
		{Name: "DESCRICAO", Type: dbf.TypeCharacter, Length: 30, Decimals: 0},
		{Name: "VALOR", Type: dbf.TypeNumeric, Length: 10, Decimals: 2},
	}

	tbl, err := dbf.CreateTable(filePath, defaultFields)
	if err != nil {
		return err
	}

	e.ActiveTable = tbl
	e.ActiveAlias = strings.ToUpper(filepath.Base(filePath))

	if e.Config.Talk {
		e.Printf("Estrutura da tabela %s criada com sucesso.\n", e.ActiveAlias)
	}
	return nil
}

func (e *Engine) cmdList(stmt *parser.Statement) error {
	if e.ActiveTable == nil {
		return fmt.Errorf("nenhum arquivo .DBF em uso (use o comando USE)")
	}

	// LIST STRUCTURE
	if stmt.Options["STRUCTURE"] == "true" {
		e.Printf("ESTRUTURA PARA O ARQUIVO: %s\n", e.ActiveTable.FilePath)
		e.Printf("NUMERO DE REGISTROS: %d\n", len(e.ActiveTable.Records))
		e.Printf("DATA DO ULTIMO UPDATE: %s\n", e.ActiveTable.Header.LastUpdate.Format("02/01/2006"))
		e.Printf("TAMANHO DO REGISTRO: %d BYTES\n\n", e.ActiveTable.Header.RecordSize)
		e.Printf("%-5s %-12s %-6s %-6s %-6s\n", "FLD", "NAME", "TYPE", "WIDTH", "DEC")
		e.Println("----------------------------------------")
		for i, f := range e.ActiveTable.Header.Fields {
			e.Printf("%03d   %-12s %-6c %-6d %-6d\n", i+1, f.Name, f.Type, f.Length, f.Decimals)
		}
		e.Printf("** TOTAL **                       %d\n", e.ActiveTable.Header.RecordSize)
		return nil
	}

	fieldsToDisplay := stmt.Fields
	if len(fieldsToDisplay) == 0 {
		fieldsToDisplay = e.ActiveTable.FieldNames()
	}

	// Cabeçalho da listagem
	e.Printf("%-6s ", "RECNO")
	for _, fn := range fieldsToDisplay {
		e.Printf("%-15s ", fn)
	}
	e.Println()
	e.Println(strings.Repeat("-", 6+(16*len(fieldsToDisplay))))

	matchCount := 0
	for i, rec := range e.ActiveTable.Records {
		recNo := i + 1

		// Filtro FOR
		if stmt.ForClause != "" && !evaluateForClause(rec, stmt.ForClause) {
			continue
		}

		delMarker := " "
		if rec.Deleted {
			delMarker = "*"
		}

		e.Printf("%05d%s ", recNo, delMarker)
		for _, fn := range fieldsToDisplay {
			val := rec.Get(fn)
			e.Printf("%-15s ", val)
		}
		e.Println()
		matchCount++
	}

	if e.Config.Talk {
		e.Printf("\n%d registros listados.\n", matchCount)
	}

	return nil
}

func (e *Engine) cmdDisplay(stmt *parser.Statement) error {
	if e.ActiveTable == nil {
		return fmt.Errorf("nenhum arquivo .DBF em uso")
	}

	if stmt.Scope == "ALL" {
		return e.cmdList(stmt)
	}

	rec := e.ActiveTable.CurrentRecord()
	if rec == nil {
		if e.Config.Talk {
			e.Println("Fim de arquivo (EOF).")
		}
		return nil
	}

	e.Printf("REGISTRO %05d:\n", e.ActiveTable.RecNo)
	for _, f := range e.ActiveTable.Header.Fields {
		e.Printf("  %-12s : %s\n", f.Name, rec.Get(f.Name))
	}
	return nil
}

func (e *Engine) cmdAppend(stmt *parser.Statement) error {
	if e.ActiveTable == nil {
		return fmt.Errorf("nenhum arquivo .DBF em uso")
	}

	rec := e.ActiveTable.AppendBlank()
	_ = e.ActiveTable.Save()

	if e.Config.Talk {
		e.Printf("Registro %d inserido.\n", len(e.ActiveTable.Records))
	}
	_ = rec
	return nil
}

func (e *Engine) cmdEdit(stmt *parser.Statement) error {
	if e.ActiveTable == nil {
		return fmt.Errorf("nenhum arquivo .DBF em uso")
	}
	if stmt.RecNo > 0 {
		if err := e.ActiveTable.Goto(stmt.RecNo); err != nil {
			return err
		}
	}
	if e.Config.Talk {
		e.Printf("Editando registro %d.\n", e.ActiveTable.RecNo)
	}
	return nil
}

func (e *Engine) cmdDelete(stmt *parser.Statement) error {
	if e.ActiveTable == nil {
		return fmt.Errorf("nenhum arquivo .DBF em uso")
	}

	count := 0
	if stmt.Scope == "ALL" {
		for _, r := range e.ActiveTable.Records {
			if stmt.ForClause == "" || evaluateForClause(r, stmt.ForClause) {
				r.Deleted = true
				count++
			}
		}
	} else {
		if rec := e.ActiveTable.CurrentRecord(); rec != nil {
			rec.Deleted = true
			count = 1
		}
	}

	_ = e.ActiveTable.Save()
	if e.Config.Talk {
		e.Printf("%d registro(s) marcado(s) para exclusão (*).\n", count)
	}
	return nil
}

func (e *Engine) cmdRecall(stmt *parser.Statement) error {
	if e.ActiveTable == nil {
		return fmt.Errorf("nenhum arquivo .DBF em uso")
	}

	count := 0
	if stmt.Scope == "ALL" {
		for _, r := range e.ActiveTable.Records {
			if stmt.ForClause == "" || evaluateForClause(r, stmt.ForClause) {
				r.Deleted = false
				count++
			}
		}
	} else {
		if rec := e.ActiveTable.CurrentRecord(); rec != nil {
			rec.Deleted = false
			count = 1
		}
	}

	_ = e.ActiveTable.Save()
	if e.Config.Talk {
		e.Printf("%d registro(s) restaurado(s).\n", count)
	}
	return nil
}

func (e *Engine) cmdPack(stmt *parser.Statement) error {
	if e.ActiveTable == nil {
		return fmt.Errorf("nenhum arquivo .DBF em uso")
	}

	removed := e.ActiveTable.Pack()
	if err := e.ActiveTable.Save(); err != nil {
		return err
	}

	if e.Config.Talk {
		e.Printf("PACK concluído: %d registros fisicamente removidos. Total restante: %d.\n", removed, len(e.ActiveTable.Records))
	}
	return nil
}

func (e *Engine) cmdZap(stmt *parser.Statement) error {
	if e.ActiveTable == nil {
		return fmt.Errorf("nenhum arquivo .DBF em uso")
	}

	e.ActiveTable.Zap()
	if err := e.ActiveTable.Save(); err != nil {
		return err
	}

	if e.Config.Talk {
		e.Println("ZAP concluído: Todos os registros foram eliminados.")
	}
	return nil
}

func (e *Engine) cmdReplace(stmt *parser.Statement) error {
	if e.ActiveTable == nil {
		return fmt.Errorf("nenhum arquivo .DBF em uso")
	}
	if stmt.Target == "" {
		return fmt.Errorf("uso: REPLACE <campo> WITH <valor> [FOR <condicao>]")
	}

	fieldName := stmt.Target
	val := stmt.WithClause
	count := 0

	if stmt.ForClause != "" {
		for _, r := range e.ActiveTable.Records {
			if evaluateForClause(r, stmt.ForClause) {
				r.Set(fieldName, val)
				count++
			}
		}
	} else {
		if rec := e.ActiveTable.CurrentRecord(); rec != nil {
			rec.Set(fieldName, val)
			count = 1
		}
	}

	_ = e.ActiveTable.Save()
	if e.Config.Talk {
		e.Printf("%d registro(s) substituído(s).\n", count)
	}
	return nil
}

func (e *Engine) cmdLocate(stmt *parser.Statement) error {
	if e.ActiveTable == nil {
		return fmt.Errorf("nenhum arquivo .DBF em uso")
	}
	if stmt.ForClause == "" {
		return fmt.Errorf("uso: LOCATE FOR <condicao>")
	}

	e.LastLocateCond = stmt.ForClause
	for i, r := range e.ActiveTable.Records {
		if evaluateForClause(r, stmt.ForClause) {
			e.ActiveTable.RecNo = i + 1
			if e.Config.Talk {
				e.Printf("REGISTRO = %05d\n", e.ActiveTable.RecNo)
			}
			return nil
		}
	}

	e.ActiveTable.RecNo = len(e.ActiveTable.Records) + 1 // EOF
	if e.Config.Talk {
		e.Println("Fim de arquivo (registro não encontrado).")
	}
	return nil
}

func (e *Engine) cmdContinue(stmt *parser.Statement) error {
	if e.ActiveTable == nil {
		return fmt.Errorf("nenhum arquivo .DBF em uso")
	}
	if e.LastLocateCond == "" {
		return fmt.Errorf("nenhum LOCATE anterior para continuar")
	}

	start := e.ActiveTable.RecNo
	for i := start; i < len(e.ActiveTable.Records); i++ {
		r := e.ActiveTable.Records[i]
		if evaluateForClause(r, e.LastLocateCond) {
			e.ActiveTable.RecNo = i + 1
			if e.Config.Talk {
				e.Printf("REGISTRO = %05d\n", e.ActiveTable.RecNo)
			}
			return nil
		}
	}

	e.ActiveTable.RecNo = len(e.ActiveTable.Records) + 1
	if e.Config.Talk {
		e.Println("Fim de arquivo (EOF).")
	}
	return nil
}

func (e *Engine) cmdCount(stmt *parser.Statement) error {
	if e.ActiveTable == nil {
		return fmt.Errorf("nenhum arquivo .DBF em uso")
	}

	count := 0
	for _, r := range e.ActiveTable.Records {
		if stmt.ForClause == "" || evaluateForClause(r, stmt.ForClause) {
			count++
		}
	}

	e.Printf("%d REGISTROS\n", count)
	return nil
}

func (e *Engine) cmdSum(stmt *parser.Statement) error {
	if e.ActiveTable == nil {
		return fmt.Errorf("nenhum arquivo .DBF em uso")
	}

	fields := stmt.Fields
	if len(fields) == 0 {
		return fmt.Errorf("especifique ao menos um campo numérico para somar")
	}

	sums := make(map[string]float64)
	count := 0

	for _, r := range e.ActiveTable.Records {
		if stmt.ForClause == "" || evaluateForClause(r, stmt.ForClause) {
			count++
			for _, fn := range fields {
				valStr := r.Get(fn)
				val, _ := strconv.ParseFloat(valStr, 64)
				sums[fn] += val
			}
		}
	}

	e.Printf("%d REGISTROS SOMADOS:\n", count)
	for _, fn := range fields {
		e.Printf("  TOTAL (%s) = %.2f\n", fn, sums[fn])
	}
	return nil
}

func (e *Engine) cmdDo(stmt *parser.Statement) error {
	if stmt.Target == "" {
		return fmt.Errorf("uso: DO <nome_script>")
	}
	path := stmt.Target
	if filepath.Ext(path) == "" {
		path += ".CMD"
	}
	return e.ExecuteScript(path)
}

func (e *Engine) cmdSet(stmt *parser.Statement) error {
	val := (stmt.SetValue == "ON" || stmt.SetValue == "TRUE" || stmt.SetValue == "1")
	switch stmt.SetParam {
	case "TALK":
		e.Config.Talk = val
	case "ECHO":
		e.Config.Echo = val
	case "BELL":
		e.Config.Bell = val
	default:
		return fmt.Errorf("parâmetro SET desconhecido: %s", stmt.SetParam)
	}
	if e.Config.Talk {
		e.Printf("SET %s %s\n", stmt.SetParam, stmt.SetValue)
	}
	return nil
}

func (e *Engine) cmdQuit(stmt *parser.Statement) error {
	if e.ActiveTable != nil {
		_ = e.ActiveTable.Save()
	}
	e.QuitRequested = true
	if e.Config.Talk {
		e.Println("*** FIM DO dBase II+ ***")
	}
	return nil
}

func (e *Engine) cmdHelp(stmt *parser.Statement) error {
	e.Println("=== COMANDOS PRINCIPAIS dBase II+ ===")
	e.Println("  USE <arquivo>               - Abre uma tabela .DBF")
	e.Println("  CREATE <arquivo>            - Cria uma nova tabela .DBF")
	e.Println("  LIST [campos] [FOR cond]    - Lista registros na tela")
	e.Println("  LIST STRUCTURE              - Exibe estrutura de campos")
	e.Println("  DISPLAY [ALL]               - Exibe dados do registro atual")
	e.Println("  APPEND [BLANK]              - Adiciona novo registro")
	e.Println("  DELETE [ALL] [FOR cond]     - Marca registro para exclusão (*)")
	e.Println("  RECALL [ALL] [FOR cond]     - Restaura registro excluído")
	e.Println("  PACK                        - Elimina fisicamente os excluídos")
	e.Println("  ZAP                         - Esvazia todos os registros")
	e.Println("  REPLACE <campo> WITH <val>  - Atualiza valores de campos")
	e.Println("  LOCATE FOR <cond>           - Busca registros por condição")
	e.Println("  CONTINUE                    - Continua última busca")
	e.Println("  COUNT [FOR cond]            - Conta registros")
	e.Println("  SUM <campo> [FOR cond]      - Totaliza campos numéricos")
	e.Println("  DO <script.cmd>             - Executa arquivo de comandos")
	e.Println("  SET TALK ON/OFF             - Ativa/desativa mensagens de status")
	e.Println("  QUIT                        - Salva alterações e encerra")
	return nil
}

// evaluateForClause avalia uma expressão simples do tipo "CAMPO = VALOR" ou "CAMPO > VALOR"
func evaluateForClause(rec *dbf.Record, expr string) bool {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return true
	}

	// Operadores comuns: =, ==, !=, <>, >, <, >=, <=, $ (substring)
	if strings.Contains(expr, "=") {
		parts := strings.SplitN(expr, "=", 2)
		fName := strings.ToUpper(strings.TrimSpace(parts[0]))
		expected := strings.Trim(strings.TrimSpace(parts[1]), "'\"")
		actual := rec.Get(fName)
		return strings.EqualFold(actual, expected)
	}

	if strings.Contains(expr, ">") {
		parts := strings.SplitN(expr, ">", 2)
		fName := strings.ToUpper(strings.TrimSpace(parts[0]))
		valStr := strings.TrimSpace(parts[1])
		expNum, err1 := strconv.ParseFloat(valStr, 64)
		actNum, err2 := strconv.ParseFloat(rec.Get(fName), 64)
		if err1 == nil && err2 == nil {
			return actNum > expNum
		}
	}

	if strings.Contains(expr, "<") {
		parts := strings.SplitN(expr, "<", 2)
		fName := strings.ToUpper(strings.TrimSpace(parts[0]))
		valStr := strings.TrimSpace(parts[1])
		expNum, err1 := strconv.ParseFloat(valStr, 64)
		actNum, err2 := strconv.ParseFloat(rec.Get(fName), 64)
		if err1 == nil && err2 == nil {
			return actNum < expNum
		}
	}

	return true
}
