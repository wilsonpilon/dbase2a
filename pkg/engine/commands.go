package engine

import (
	"fmt"
	"path/filepath"
	"strings"

	"dbase2a/pkg/dbf"
	"dbase2a/pkg/eval"
	"dbase2a/pkg/help"
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
	case parser.CmdBrowse:
		return e.cmdBrowse(stmt)
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
	case parser.CmdStore:
		return e.cmdStore(stmt)
	case parser.CmdPrint:
		return e.cmdPrint(stmt, true)
	case parser.CmdPrintSame:
		return e.cmdPrint(stmt, false)
	case parser.CmdGoto:
		return e.cmdGoto(stmt)
	case parser.CmdSkip:
		return e.cmdSkip(stmt)
	case parser.CmdDo:
		return e.cmdDo(stmt)
	case parser.CmdSet:
		return e.cmdSet(stmt)
	case parser.CmdQuit:
		return e.cmdQuit(stmt)
	case parser.CmdHelp:
		return e.cmdHelp(stmt)
	case parser.CmdClear:
		return e.cmdClear(stmt)
	default:
		return fmt.Errorf("comando não reconhecido ou sintaxe inválida: '%s'", stmt.Raw)
	}
}

func (e *Engine) cmdUse(stmt *parser.Statement) error {
	if stmt.Target == "" {
		if e.ActiveTable != nil {
			_ = e.ActiveTable.Close()
			e.ActiveTable = nil
			e.ActiveAlias = ""
			e.CurrentRecordIndex = 0
		}
		return nil
	}

	filePath := stmt.Target
	if !strings.HasSuffix(strings.ToUpper(filePath), ".DBF") {
		filePath += ".DBF"
	}

	tbl, err := dbf.OpenTable(filePath)
	if err != nil {
		return fmt.Errorf("erro ao abrir tabela '%s': %w", filePath, err)
	}

	if e.ActiveTable != nil {
		_ = e.ActiveTable.Close()
	}

	e.ActiveTable = tbl
	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	e.ActiveAlias = strings.ToUpper(strings.TrimSuffix(base, ext))
	e.CurrentRecordIndex = 1

	if e.Config.Talk {
		e.Printf("Tabela ativa: %s (%d registros)\n", e.ActiveAlias, len(tbl.Records))
	}
	return nil
}

func (e *Engine) cmdCreate(stmt *parser.Statement) error {
	return e.InteractiveCreate(stmt.Target)
}

func (e *Engine) cmdList(stmt *parser.Statement) error {
	return e.displayOrList(stmt, true)
}

func (e *Engine) cmdDisplay(stmt *parser.Statement) error {
	return e.displayOrList(stmt, false)
}

func (e *Engine) displayOrList(stmt *parser.Statement, isList bool) error {
	if e.ActiveTable == nil {
		return fmt.Errorf("nenhuma tabela aberta. Utilize 'USE <arquivo>' primeiro")
	}

	if stmt.Options["STRUCTURE"] == "true" {
		e.Printf("ESTRUTURA DA TABELA: %s\n", e.ActiveAlias)
		e.Printf("NUMERO DE REGISTROS: %05d\n", len(e.ActiveTable.Records))
		e.Printf("CAMPOS:\n")
		e.Printf("  NOME        TIPO  TAM  DEC\n")
		for _, f := range e.ActiveTable.Header.Fields {
			e.Printf("  %-10s   %c   %3d  %3d\n", f.Name, f.Type, f.Length, f.Decimals)
		}
		return nil
	}

	if len(e.ActiveTable.Records) == 0 {
		if e.Config.Talk {
			e.Println("Tabela vazia.")
		}
		return nil
	}

	// Cabeçalho de colunas
	var headerCols []string
	headerCols = append(headerCols, "REC# ")
	fieldsToShow := stmt.Fields
	if len(fieldsToShow) == 0 {
		for _, f := range e.ActiveTable.Header.Fields {
			fieldsToShow = append(fieldsToShow, f.Name)
		}
	}

	for _, fn := range fieldsToShow {
		headerCols = append(headerCols, fmt.Sprintf("%-15s", fn))
	}
	e.Println(strings.Join(headerCols, " "))

	startIdx := 0
	endIdx := len(e.ActiveTable.Records)

	if !isList && stmt.Scope != "ALL" {
		// DISPLAY sem ALL exibe o registro atual
		curr := e.CurrentRecordIndex - 1
		if curr < 0 {
			curr = 0
		}
		if curr >= len(e.ActiveTable.Records) {
			e.Println("FIM DE ARQUIVO ALCANCADO")
			return nil
		}
		startIdx = curr
		endIdx = curr + 1
	}

	for i := startIdx; i < endIdx; i++ {
		rec := e.ActiveTable.Records[i]
		recNo := i + 1

		// Avalia cláusula WHILE (se falhar, interrompe loop)
		if stmt.WhileClause != "" {
			passWhile, err := e.evaluateCondition(rec, recNo, stmt.WhileClause)
			if err != nil {
				return err
			}
			if !passWhile {
				break
			}
		}

		// Avalia cláusula FOR (se falhar, pula registro)
		if stmt.ForClause != "" {
			passFor, err := e.evaluateCondition(rec, recNo, stmt.ForClause)
			if err != nil {
				return err
			}
			if !passFor {
				continue
			}
		}

		delMark := " "
		if rec.IsDeleted() {
			delMark = "*"
		}

		var rowCols []string
		rowCols = append(rowCols, fmt.Sprintf("%05d%s", recNo, delMark))

		for _, fn := range fieldsToShow {
			val := rec.Get(fn)
			if val == "" {
				// Tenta avaliar como expressão se não for campo direto
				ctx := e.GetEvalContext(rec, recNo)
				if v, err := eval.Eval(fn, ctx); err == nil && v.Type != eval.TypeNull {
					val = v.AsString()
				}
			}
			rowCols = append(rowCols, fmt.Sprintf("%-15s", val))
		}
		e.Println(strings.Join(rowCols, " "))
		e.CurrentRecordIndex = recNo
	}

	return nil
}

func (e *Engine) cmdAppend(stmt *parser.Statement) error {
	if e.ActiveTable == nil {
		return fmt.Errorf("nenhuma tabela aberta para APPEND")
	}

	if stmt.Options["BLANK"] == "true" {
		_ = e.ActiveTable.AppendBlank()
		e.CurrentRecordIndex = len(e.ActiveTable.Records)

		if err := e.ActiveTable.Save(); err != nil {
			return fmt.Errorf("erro ao salvar novo registro: %w", err)
		}

		if e.Config.Talk {
			e.Printf("Registro %d adicionado em branco.\n", len(e.ActiveTable.Records))
		}
		return nil
	}

	return e.InteractiveAppend()
}

func (e *Engine) cmdEdit(stmt *parser.Statement) error {
	if e.ActiveTable == nil {
		return fmt.Errorf("nenhuma tabela aberta para EDIT")
	}

	recNo := stmt.RecNo
	if recNo <= 0 {
		recNo = e.CurrentRecordIndex
	}

	return e.InteractiveEdit(recNo)
}

func (e *Engine) cmdBrowse(stmt *parser.Statement) error {
	return e.cmdList(&parser.Statement{Type: parser.CmdList, Scope: "ALL"})
}

func (e *Engine) cmdDelete(stmt *parser.Statement) error {
	if e.ActiveTable == nil {
		return fmt.Errorf("nenhuma tabela aberta")
	}

	count := 0
	startIdx := 0
	endIdx := len(e.ActiveTable.Records)

	if stmt.Scope != "ALL" && stmt.ForClause == "" && stmt.WhileClause == "" {
		curr := e.CurrentRecordIndex - 1
		if curr >= 0 && curr < len(e.ActiveTable.Records) {
			startIdx = curr
			endIdx = curr + 1
		}
	}

	for i := startIdx; i < endIdx; i++ {
		rec := e.ActiveTable.Records[i]
		recNo := i + 1

		if stmt.WhileClause != "" {
			passWhile, err := e.evaluateCondition(rec, recNo, stmt.WhileClause)
			if err != nil {
				return err
			}
			if !passWhile {
				break
			}
		}

		if stmt.ForClause != "" {
			passFor, err := e.evaluateCondition(rec, recNo, stmt.ForClause)
			if err != nil {
				return err
			}
			if !passFor {
				continue
			}
		}

		rec.Delete()
		count++
	}

	_ = e.ActiveTable.Save()
	if e.Config.Talk {
		e.Printf("%05d REGISTRO(S) MARCADOS PARA EXCLUSAO\n", count)
	}
	return nil
}

func (e *Engine) cmdRecall(stmt *parser.Statement) error {
	if e.ActiveTable == nil {
		return fmt.Errorf("nenhuma tabela aberta")
	}

	count := 0
	startIdx := 0
	endIdx := len(e.ActiveTable.Records)

	if stmt.Scope != "ALL" && stmt.ForClause == "" && stmt.WhileClause == "" {
		curr := e.CurrentRecordIndex - 1
		if curr >= 0 && curr < len(e.ActiveTable.Records) {
			startIdx = curr
			endIdx = curr + 1
		}
	}

	for i := startIdx; i < endIdx; i++ {
		rec := e.ActiveTable.Records[i]
		recNo := i + 1

		if stmt.WhileClause != "" {
			passWhile, err := e.evaluateCondition(rec, recNo, stmt.WhileClause)
			if err != nil {
				return err
			}
			if !passWhile {
				break
			}
		}

		if stmt.ForClause != "" {
			passFor, err := e.evaluateCondition(rec, recNo, stmt.ForClause)
			if err != nil {
				return err
			}
			if !passFor {
				continue
			}
		}

		if rec.IsDeleted() {
			rec.Recall()
			count++
		}
	}

	_ = e.ActiveTable.Save()
	if e.Config.Talk {
		e.Printf("%05d REGISTRO(S) RECUPERADOS\n", count)
	}
	return nil
}

func (e *Engine) cmdPack(stmt *parser.Statement) error {
	if e.ActiveTable == nil {
		return fmt.Errorf("nenhuma tabela aberta")
	}

	removed := e.ActiveTable.Pack()
	if err := e.ActiveTable.Save(); err != nil {
		return fmt.Errorf("erro ao salvar após PACK: %w", err)
	}

	if e.CurrentRecordIndex > len(e.ActiveTable.Records) {
		e.CurrentRecordIndex = len(e.ActiveTable.Records)
	}

	if e.Config.Talk {
		e.Printf("PACK CONCLUIDO: %d registro(s) excluido(s) permanentemente. Restam %d.\n", removed, len(e.ActiveTable.Records))
	}
	return nil
}

func (e *Engine) cmdZap(stmt *parser.Statement) error {
	if e.ActiveTable == nil {
		return fmt.Errorf("nenhuma tabela aberta")
	}

	e.ActiveTable.Zap()
	if err := e.ActiveTable.Save(); err != nil {
		return fmt.Errorf("erro ao salvar após ZAP: %w", err)
	}

	e.CurrentRecordIndex = 0
	if e.Config.Talk {
		e.Println("ZAP CONCLUIDO: Todos os registros foram apagados.")
	}
	return nil
}

func (e *Engine) cmdReplace(stmt *parser.Statement) error {
	if e.ActiveTable == nil {
		return fmt.Errorf("nenhuma tabela aberta")
	}

	fieldName := stmt.Target
	if fieldName == "" {
		return fmt.Errorf("sintaxe: REPLACE [scope] <campo> WITH <expressao> [FOR <cond>] [WHILE <cond>]")
	}

	count := 0
	startIdx := 0
	endIdx := len(e.ActiveTable.Records)

	if stmt.Scope != "ALL" && stmt.ForClause == "" && stmt.WhileClause == "" {
		curr := e.CurrentRecordIndex - 1
		if curr >= 0 && curr < len(e.ActiveTable.Records) {
			startIdx = curr
			endIdx = curr + 1
		}
	}

	for i := startIdx; i < endIdx; i++ {
		rec := e.ActiveTable.Records[i]
		recNo := i + 1

		if stmt.WhileClause != "" {
			passWhile, err := e.evaluateCondition(rec, recNo, stmt.WhileClause)
			if err != nil {
				return err
			}
			if !passWhile {
				break
			}
		}

		if stmt.ForClause != "" {
			passFor, err := e.evaluateCondition(rec, recNo, stmt.ForClause)
			if err != nil {
				return err
			}
			if !passFor {
				continue
			}
		}

		ctx := e.GetEvalContext(rec, recNo)
		val, err := eval.Eval(stmt.WithClause, ctx)
		if err != nil {
			return fmt.Errorf("erro ao avaliar cláusula WITH '%s': %w", stmt.WithClause, err)
		}

		rec.Set(fieldName, val.AsString())
		count++
		e.CurrentRecordIndex = recNo
	}

	_ = e.ActiveTable.Save()
	if e.Config.Talk {
		e.Printf("%05d REGISTRO(S) SUBSTITUIDO(S)\n", count)
	}
	return nil
}

func (e *Engine) cmdLocate(stmt *parser.Statement) error {
	if e.ActiveTable == nil {
		return fmt.Errorf("nenhuma tabela aberta")
	}

	e.LastLocateCond = stmt.ForClause
	return e.findNextLocate(0)
}

func (e *Engine) cmdContinue(stmt *parser.Statement) error {
	if e.ActiveTable == nil {
		return fmt.Errorf("nenhuma tabela aberta")
	}
	if e.LastLocateCond == "" {
		return fmt.Errorf("nenhum comando LOCATE anterior ativo")
	}

	start := e.CurrentRecordIndex
	return e.findNextLocate(start)
}

func (e *Engine) findNextLocate(startIdx int) error {
	for i := startIdx; i < len(e.ActiveTable.Records); i++ {
		rec := e.ActiveTable.Records[i]
		recNo := i + 1

		pass, err := e.evaluateCondition(rec, recNo, e.LastLocateCond)
		if err != nil {
			return err
		}

		if pass {
			e.CurrentRecordIndex = recNo
			if e.Config.Talk {
				e.Printf("REGISTRO = %05d\n", recNo)
			}
			return nil
		}
	}

	e.CurrentRecordIndex = len(e.ActiveTable.Records) + 1
	if e.Config.Talk {
		e.Println("FIM DE ARQUIVO ALCANCADO")
	}
	return nil
}

func (e *Engine) cmdCount(stmt *parser.Statement) error {
	if e.ActiveTable == nil {
		return fmt.Errorf("nenhuma tabela aberta")
	}

	count := 0
	for i, rec := range e.ActiveTable.Records {
		recNo := i + 1

		if stmt.WhileClause != "" {
			passWhile, err := e.evaluateCondition(rec, recNo, stmt.WhileClause)
			if err != nil {
				return err
			}
			if !passWhile {
				break
			}
		}

		if stmt.ForClause != "" {
			passFor, err := e.evaluateCondition(rec, recNo, stmt.ForClause)
			if err != nil {
				return err
			}
			if !passFor {
				continue
			}
		}

		count++
	}

	if stmt.Target != "" {
		e.Variables[strings.ToUpper(stmt.Target)] = eval.NewNumberValue(float64(count))
	}

	e.Printf("%05d REGISTROS\n", count)
	return nil
}

func (e *Engine) cmdSum(stmt *parser.Statement) error {
	if e.ActiveTable == nil {
		return fmt.Errorf("nenhuma tabela aberta")
	}

	fields := stmt.Fields
	if len(fields) == 0 {
		return fmt.Errorf("sintaxe: SUM <campo1> [, <campo2>] [FOR <cond>] [WHILE <cond>] [TO <var>]")
	}

	totals := make([]float64, len(fields))

	for i, rec := range e.ActiveTable.Records {
		recNo := i + 1

		if stmt.WhileClause != "" {
			passWhile, err := e.evaluateCondition(rec, recNo, stmt.WhileClause)
			if err != nil {
				return err
			}
			if !passWhile {
				break
			}
		}

		if stmt.ForClause != "" {
			passFor, err := e.evaluateCondition(rec, recNo, stmt.ForClause)
			if err != nil {
				return err
			}
			if !passFor {
				continue
			}
		}

		ctx := e.GetEvalContext(rec, recNo)
		for fIdx, fName := range fields {
			val, err := eval.Eval(fName, ctx)
			if err == nil {
				if num, err := val.AsNumber(); err == nil {
					totals[fIdx] += num
				}
			}
		}
	}

	if stmt.Target != "" && len(totals) > 0 {
		e.Variables[strings.ToUpper(stmt.Target)] = eval.NewNumberValue(totals[0])
	}

	var parts []string
	for fIdx, fName := range fields {
		parts = append(parts, fmt.Sprintf("%s: %.2f", fName, totals[fIdx]))
	}
	e.Printf("SOMA: %s\n", strings.Join(parts, " | "))
	return nil
}

func (e *Engine) cmdStore(stmt *parser.Statement) error {
	if stmt.Target == "" {
		return fmt.Errorf("sintaxe: STORE <expressao> TO <variavel>")
	}

	ctx := e.GetEvalContext(e.getCurrentRecord(), e.CurrentRecordIndex)
	val, err := eval.Eval(stmt.Expression, ctx)
	if err != nil {
		return fmt.Errorf("erro ao avaliar expressão STORE: %w", err)
	}

	e.Variables[strings.ToUpper(stmt.Target)] = val

	if e.Config.Talk {
		e.Printf("%s = %s\n", strings.ToUpper(stmt.Target), val.AsString())
	}
	return nil
}

func (e *Engine) cmdPrint(stmt *parser.Statement, addNewline bool) error {
	ctx := e.GetEvalContext(e.getCurrentRecord(), e.CurrentRecordIndex)
	val, err := eval.Eval(stmt.Expression, ctx)
	if err != nil {
		return fmt.Errorf("erro na avaliação de '?' : %w", err)
	}

	if addNewline {
		e.Println(val.AsString())
	} else {
		e.Printf("%s", val.AsString())
	}
	return nil
}

func (e *Engine) cmdGoto(stmt *parser.Statement) error {
	if e.ActiveTable == nil {
		return fmt.Errorf("nenhuma tabela aberta")
	}

	if strings.ToUpper(stmt.Target) == "TOP" {
		e.CurrentRecordIndex = 1
	} else if strings.ToUpper(stmt.Target) == "BOTTOM" {
		e.CurrentRecordIndex = len(e.ActiveTable.Records)
	} else if stmt.RecNo > 0 {
		if stmt.RecNo > len(e.ActiveTable.Records) {
			e.CurrentRecordIndex = len(e.ActiveTable.Records) + 1
		} else {
			e.CurrentRecordIndex = stmt.RecNo
		}
	}

	if e.Config.Talk {
		if e.CurrentRecordIndex > len(e.ActiveTable.Records) {
			e.Println("FIM DE ARQUIVO ALCANCADO")
		} else {
			e.Printf("REGISTRO ATUAL: %05d\n", e.CurrentRecordIndex)
		}
	}
	return nil
}

func (e *Engine) cmdSkip(stmt *parser.Statement) error {
	if e.ActiveTable == nil {
		return fmt.Errorf("nenhuma tabela aberta")
	}

	delta := stmt.RecNo
	if delta == 0 {
		delta = 1
	}

	e.CurrentRecordIndex += delta
	if e.CurrentRecordIndex < 1 {
		e.CurrentRecordIndex = 1
	}

	if e.Config.Talk {
		if e.CurrentRecordIndex > len(e.ActiveTable.Records) {
			e.Println("FIM DE ARQUIVO ALCANCADO")
		} else {
			e.Printf("REGISTRO ATUAL: %05d\n", e.CurrentRecordIndex)
		}
	}
	return nil
}

func (e *Engine) cmdDo(stmt *parser.Statement) error {
	if stmt.Target == "" {
		return fmt.Errorf("sintaxe: DO <nome_script>")
	}
	return e.ExecuteScript(stmt.Target)
}

func (e *Engine) cmdSet(stmt *parser.Statement) error {
	param := stmt.SetParam
	val := stmt.SetValue

	switch param {
	case "TALK":
		e.Config.Talk = (val == "ON")
	case "ECHO":
		e.Config.Echo = (val == "ON")
	case "EXACT":
		e.Config.Exact = (val == "ON")
	case "BELL":
		e.Config.Bell = (val == "ON")
	default:
		if e.Config.Talk {
			e.Printf("SET %s %s aceito.\n", param, val)
		}
	}
	return nil
}

func (e *Engine) cmdQuit(stmt *parser.Statement) error {
	if e.ActiveTable != nil {
		_ = e.ActiveTable.Close()
		e.ActiveTable = nil
	}
	e.QuitRequested = true
	if e.Config.Talk {
		e.Println("*** FIM DO DBASE II++ ***")
	}
	return nil
}

func (e *Engine) cmdHelp(stmt *parser.Statement) error {
	if e.Help == nil {
		e.Help = help.NewHelpSystem()
		_ = e.Help.LoadDefault()
	}

	pager := help.NewPager(nil, e.Out)

	target := strings.TrimSpace(stmt.Target)
	if target == "" {
		// HELP sem argumentos mostra a introdução e guia de uso do DBASEMSG.TXT
		intro, ok := e.Help.GetTopic("DBASE")
		if !ok {
			intro, ok = e.Help.GetTopic("HELP")
		}
		if ok {
			return pager.Display(intro)
		} else {
			e.Println("Não há mensagem HELP.")
		}
		return nil
	}

	topicText, ok := e.Help.GetTopic(target)
	if !ok {
		e.Printf("Não há mensagem HELP para '%s'.\n", target)
		e.Println("Digite 'HELP' ou 'HELP COMMANDS' para ver a lista de comandos e tópicos disponíveis.")
		return nil
	}

	return pager.Display(topicText)
}

func (e *Engine) cmdClear(stmt *parser.Statement) error {
	e.Printf("\033[2J\033[H")
	return nil
}

func (e *Engine) getCurrentRecord() *dbf.Record {
	if e.ActiveTable == nil || len(e.ActiveTable.Records) == 0 {
		return nil
	}
	curr := e.CurrentRecordIndex - 1
	if curr < 0 {
		curr = 0
	}
	if curr >= len(e.ActiveTable.Records) {
		return nil
	}
	return e.ActiveTable.Records[curr]
}

func (e *Engine) evaluateCondition(rec *dbf.Record, recNo int, condStr string) (bool, error) {
	ctx := e.GetEvalContext(rec, recNo)
	return eval.EvalBool(condStr, ctx)
}
