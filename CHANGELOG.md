# Changelog

Todas as alterações notáveis neste projeto serão documentadas neste arquivo.

O formato é baseado em [Keep a Changelog](https://keepachangelog.com/pt-BR/1.0.0/),
e este projeto adota versionamento no padrão `vX.Y.Z` onde:
- `X` = Pacote de funcionalidades maiores / fases arquiteturais
- `Y` = Novo recurso (feature) introduzido ou finalizado
- `Z` = Contador de compilação/build incrementado automaticamente a cada build
- Cada lançamento recebe o codinome de um clássico filme de terror/horror.

---

## [v0.6.4] - 2026-10-08 - *The Exorcist*

### Adicionado
- **Documento de Handover & Transição (`HANDOVER.md`)**:
  - Guia completo de arquitetura, estado de código e instruções detalhadas para continuação imediata do projeto em outro computador utilizando GoLand e Junie.
  - Registro detalhado das decisões de design, status dos módulos e roadmap dos passos seguintes.
- **Suporte a Flag `--no-increment` / `-NoIncrement` nos Scripts de Build**:
  - Possibilidade de regerar os pacotes ZIP (Windows) e TAR.GZ (Linux) mantendo rigorosamente a mesma versão semântica e hash de compilação.
- **Renomeação Integral para dBase II++ (*dBase II Plus Plus*)**:
  - Atualização de todos os pacotes, banners, CLI, TUI, GUI, geradores MSX e documentação viva.

### Modificado
- **Criação Interativa de Tabelas (`CREATE`) (`pkg/engine/interactive.go`)**:
  - Solicitação interativa do nome do arquivo quando omitido (`ENTER FILENAME: `).
  - Exibição do cabeçalho clássico: `ENTER RECORD STRUCTURE AS FOLLOWS: FIELD  NAME,TYPE,WIDTH,DECIMAL PLACES`.
  - Prompt linha a linha sequencial (`001    `, `002    `, etc.) para definição de campos.
  - Suporte a múltiplos formatos de entrada (`NOME, C, 30`, `IDADE, N, 3`, `SALARIO, N, 10, 2`, `ATIVO, L` ou delimitado por espaços).
  - Validação estrita com repetição do campo em caso de erro: tipos `C`, `N`, `L`, limites de tamanho (1-254), casas decimais e prevenção de campos duplicados.
  - Finalização da estrutura ao pressionar `Enter` em branco e criação imediata do arquivo `.DBF` binário compatível.
  - Pergunta clássica para inclusão imediata: `INPUT DATA NOW? (Y/N)` com transição direta ao modo `APPEND`.
- **Modo Interativo de Inclusão (`APPEND`) e Edição (`EDIT`)**:
  - Tela simples e intuitiva de inserção de dados campo a campo com indicação de tamanho e valores atuais.
  - Encerramento automático do `APPEND` ao pressionar `Enter` no primeiro campo de um registro em branco, salvando a tabela e retornando ao prompt.
  - Suporte a `APPEND BLANK` para scripts não-interativos.
  - Edição interativa de registros com `EDIT [recno]` preservando valores não alterados.
- **Mapeamento e Documentação de Teclas e Atalhos**:
  - Suporte a comandos de teclado clássicos do dBASE II: `Ctrl+W` (salvar e sair), `Ctrl+Q` (cancelar/sair), `Ctrl+U` (alternar exclusão lógica), `Ctrl+E`/`Ctrl+X` (navegar campos), `Ctrl+C`/`Ctrl+R` (navegar registros).
  - Suporte a teclas modernas de terminal: `Enter`, `Esc`, `Setas Cima/Baixo`, `PageUp/PageDown`, `Tab`/`Shift+Tab`.
  - Documentação detalhada em `SPEC.md`, `README.md` e `RELEASE.md`.

## [v0.5.0] - 2026-10-08 - *The Thing*

### Adicionado
- **Paginador Interativo no Subsistema de Ajuda (`pkg/help/pager.go`)**:
  - Exibição com rolagem de páginas e linhas para textos longos de documentação em terminais interativos.
  - Navegação bidirecional por teclado: `[Espaço]` / `[PageDown]` (avançar tela), `[b]` / `[PageUp]` (voltar tela), `[Setas/j/k]` (rolar linhas), `[g/Home]` (início), `[G/End]` (fim) e `[q/Esc]` (sair).
  - Barra de status com posição em linhas e porcentagem lida.
- **Comando `HELP COMMANDS` / `HELP COMANDOS`**:
  - Catálogo sintetizado de todos os comandos do dBASE II com resumo de uma linha de sua finalidade.
- **Editor de Linha CLI com Autocomplete Cíclico por TAB (`pkg/ui/cli/autocomplete.go` e `pkg/ui/cli/line_editor.go`)**:
  - Completa comandos com único match automaticamente ou circula entre múltiplos candidatos a cada pressão sucessiva da tecla TAB.
  - Autocomplete sensível ao contexto: comandos dBASE, tópicos do HELP, parâmetros de SET e arquivos `.DBF`/`.CMD` no disco.
- **Realce de Sintaxe Dinâmico (Syntax Highlighting) & Erros em Vermelho (`pkg/ui/cli/syntax.go`)**:
  - Destaque em tempo real no prompt `.` com cores ANSI: Comandos (Ciano), Cláusulas (Branco Negrito), Operadores (Amarelo), Funções (Magenta), Nomes de Arquivos (Verde), Strings (Verde Claro), Números (Azul), Comentários (Cinza).
  - Mensagens de erro padronizadas e destacadas em Vermelho Brilhante (`*** ERRO:`).

## [v0.4.0] - 2026-10-08 - *Halloween*

### Adicionado
- **Subsistema de Ajuda Interativa (`pkg/help`)**:
  - Implementação completa do leitor e indexador do arquivo clássico `DBASEMSG.TXT` (Ajuda Oficial do dBASE II Plus MSX).
  - Conversor transparente da ROM de caracteres de 8-bits do MSX nacional (Gradiente Expert / Sharp Hotbit) para UTF-8 no PC, corrigindo toda a acentuação e cedilhas.
  - Suporte a busca flexível e tolerante de tópicos: maiúsculas/minúsculas, com ou sem acentos, com ou sem hífens (ex: `HELP FUNÇÕES` ou `HELP FUNCOES`, `HELP TELA-CHEIA` ou `HELP TELA CHEIA`).
  - Integração dos manuais do editor de tela MSX-DOS (`MED.HLP` e `MED.MES`) nos tópicos `HELP EDITOR` e `HELP MED`.
  - Mais de 86 tópicos indexados incluindo comandos (`CREATE`, `USE`, `LIST`, `APPEND`, `REPLACE`, `STORE`, `?`, `??`, `@`), utilitários (`ZIP.COM`), mensagens de erro (`ERROS`), limites de sistema (`LIMITES`) e funções (`FUNÇÕES`).
- **Comando HELP no Motor Relacional (`pkg/engine`)**:
  - `HELP` (sem parâmetros): exibe a introdução do dBASE II Plus, instruções de uso e arquivos do sistema.
  - `HELP <comando/tópico>`: exibe a sintaxe, parâmetros e exemplos detalhados do comando informado.
  - Integração com todas as interfaces: CLI (`.`), TUI, GUI e execução em lote (`--exec "HELP ..."`).

## [v0.3.0] - 2026-10-08 - *Psycho*

### Adicionado
- **Integração Avançada do MSXgl (`pkg/msx`)**:
  - Geração nativa de templates C utilizando as APIs do MSXgl (`msxgl.h`, `dos.h`, `dos_mapper.h`, `vdp.h`, `print.h`, `string.h`).
  - Geração automática de `msxgl_config.h` e `project_config.js` pré-configurados para MSX 2+ (`MSX_22P`), MSX-DOS 2 (`TARGET_DOS2`) e modo texto de 80 colunas (`VDP_MODE_TEXT1`).
  - Suporte ao gerenciamento de **Memory Mapper** (256 KB) via rotinas `DOSMapper_Alloc` e `DOSMapper_SetPage2` para bufferização rápida de tabelas.
  - Implementação de navegador interativo de registros de 80 colunas com comandos de navegação (`[N]ext`, `[P]rev`, `[G]oto`, `[L]ist`, `[Q]uit`).
- **Conversor Intel HEX para MSX-DOS COM em Go Puro (`pkg/msx/hex2bin.go`)**:
  - Decodificação de registros Intel HEX (.ihx) gerados pelo SDCC e alinhamento em 0x0100 (TPA do MSX-DOS) sem dependências externas.
  - Flag de linha de comando `--hex2com <arquivo.ihx> [--out <arquivo.com>]`.
- **Automação de Compilação e Exportação MSX**:
  - Flags de linha de comando: `--export-msx <tabela.dbf>`, `--compile-msx <tabela.dbf>`, `--sdk msxgl`, `--mapper <tamanho_kb>`.
  - Geração de `Makefile`, `build.bat` e `build.sh` integrados em cada projeto exportado.

## [v0.2.0] - 2026-10-08 - *The Shining*

### Adicionado
- **Avaliador de Expressões dBASE II (`pkg/eval`)**:
  - Motor léxico e sintático recursivo descendente com precedência matemática e lógica estrita.
  - Tipagem polimórfica completa: Caractere (`C`), Numérico (`N`), Lógico (`L`) e Nulo.
  - Operadores relacionais: `=`, `<>`, `!=`, `#`, `<`, `>`, `<=`, `>=`.
  - Operador de substring clássico do dBASE: `$` (ex: `"MSX" $ DESCRICAO`).
  - Operadores lógicos com curto-circuito: `.AND.`, `.OR.`, `.NOT.`, `!`.
  - Operadores aritméticos com suporte a precedência e parênteses: `+`, `-`, `*`, `/`, unários `+` e `-`.
  - Funções intrínsecas: `TRIM`, `RTRIM`, `LTRIM`, `ALLTRIM`, `UPPER`, `LOWER`, `LEN`, `SUBSTR`, `VAL`, `STR`, `INT`, `ABS`, `DELETED()`, `RECNO()`, `EOF()`, `DATE()`, `TIME()`, `IIF()`.
- **Filtros Condicionais no Motor de Comandos (`pkg/engine`)**:
  - Cláusula `FOR <expressao>` com avaliação dinâmica campo a campo em `LIST`, `DISPLAY`, `COUNT`, `SUM`, `LOCATE`, `REPLACE`, `DELETE`, `RECALL`.
  - Cláusula `WHILE <expressao>` para interrupção de fluxo quando a condição de continuidade se torna falsa.
  - Avaliação de expressões na cláusula `WITH` do comando `REPLACE` (ex: `REPLACE PRECO WITH PRECO * 1.10 FOR ATIVO = .T.`).
  - Suporte a variáveis de memória com `STORE <expressao> TO <var>` e atribuição direta `<var> = <expressao>`.
  - Comandos de impressão e avaliação imediata de expressões: `? <expressao>` e `?? <expressao>`.
  - Comandos de navegação de registros: `GOTO <n>`, `GO TOP`, `GO BOTTOM`, `SKIP [n]`.
  - Opção `TO <var>` nos comandos `COUNT` e `SUM` para captura de totais em variáveis de memória.

## [v0.1.0] - 2026-10-08 - *Nosferatu*

### Adicionado
- **Documentação e Arquitetura**:
  - `SPEC.md`: Documento vivo contendo a especificação detalhada da arquitetura, formato DBF, comandos dBASE II, pipeline MSX e roadmap de fases.
  - `README.md`: Apresentação geral do projeto, guia de instalação, modos de execução (GUI, TUI, CLI) e parâmetros de linha de comando.
  - `RELEASE.md`: Histórico de entregas, notas de lançamento e codinomes temáticos.
  - `LICENSE`: Licenciamento oficial GNU General Public License v3.0 (GPL-3.0).
- **Core / DBF Engine**:
  - Parser e gravador de tabelas dBASE II (`.DBF`) compatível com cabeçalho `0x02` de 521 bytes e descritores de campos de 16 bytes.
  - Suporte completo a tipos de dados originais: Caractere (`C`), Numérico (`N`) e Lógico (`L`).
  - Suporte ao marcador de registros excluídos (`*` e espaço) e terminação com `0x1A` (EOF/Ctrl+Z).
- **Persistência SQLite**:
  - Camada de armazenamento e configurações de usuário, histórico e metadados em SQLite (`dbase2plus.db`).
- **Interfaces de Usuário Multiplataforma**:
  - **CLI**: Prompt interativo do dBase II (`.`) com suporte a execução direta de comandos (`USE`, `LIST`, `APPEND`, `PACK`, `LOCATE`, etc.).
  - **TUI**: Interface em modo texto formatada com tabelas e caixas para terminais ANSI.
  - **GUI**: Interface desktop moderna desenvolvida em Fyne com visualização de dados, console de comandos e painel MSX.
- **Pipeline MSX (MSX 2+ / MSX-DOS 2 / 256KB)**:
  - Gerador de código fonte C/Assembly integrado aos SDKs Fusion-C (MSX Fusion) e MSXgl.
  - Geração de templates de código compatíveis com o compilador SDCC para geração de binários `.COM`.
- **Automação de Build Multiplataforma**:
  - `build.ps1`: Script PowerShell para compilação no Windows, resolução de dependências, incremento automático de $Z$, geração de pacote ZIP e população da pasta `release/`.
  - `build.sh`: Script Bash para compilação no Linux/WSL, geração de pacote TAR.GZ e preparação da pasta `release/`.
