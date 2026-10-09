 # dBase II++ — Documento de Transição e Continuidade (HANDOVER.md)

> **Para uso com JetBrains GoLand e Junie (AI Agent)**  
> Este documento foi elaborado para permitir a transição imediata do projeto para um novo computador, preservando todo o contexto arquitetural, histórico de decisões, estado atual do código e próximos passos do roadmap.

---

## 📌 1. Visão Geral e Identidade do Projeto

* **Nome Oficial**: **dBase II++** (*dBase II Plus Plus*) — Versão estendida e modernizada a partir do clássico dBASE II+ do MSX.
* **Linguagem Principal**: **Go (Golang 1.22+)**.
* **Banco de Dados de Configurações/Metadados**: **SQLite embutido** (`modernc.org/sqlite`, sem dependência de DLLs externas CGO).
* **Interface Gráfica**: **Fyne** (`pkg/ui/gui`).
* **Interface de Terminal**: **TUI** (`pkg/ui/tui`) e **CLI Interativa** (`pkg/ui/cli`) com prompt clássico de ponto `.`.
* **Ecossistema MSX**: Geração de código C moderno para **MSX 2+** com **256 KB Memory Mapper** e **MSX-DOS 2**, utilizando a framework **MSXgl** e o compilador **SDCC**.
* **Licença**: **GNU General Public License v3.0 (GPL-3.0)**.
* **Versionamento**: Semântico `vX.Y.Z` com codinomes de clássicos do cinema de terror (*The Exorcist*, *The Thing*, *Halloween*, etc.). Versão atual: **`v0.6.4`**.

---

## 🏗️ 2. Arquitetura e Estrutura de Diretórios

```
C:\dos\dbase2a\ (ou raiz do projeto)
├── cmd/
│   └── dbase2plus/             # Ponto de entrada unificado da aplicação (main.go)
├── pkg/
│   ├── config/                 # Gerenciamento de configurações globais e caminhos
│   ├── dbf/                    # Motor binário do formato dBASE II (.DBF - 521 bytes, cabeçalho 0x02)
│   ├── engine/                 # Motor de execução relacional, comandos e rotinas interativas
│   │   ├── commands.go         # Execução dos comandos (CREATE, USE, LIST, APPEND, DELETE, etc.)
│   │   ├── interactive.go      # Telas interativas: CREATE passo a passo, APPEND e EDIT
│   │   └── engine.go           # Estrutura de estado Engine e ciclo de vida
│   ├── eval/                   # Avaliador de expressões polimórfico, funções e variáveis
│   ├── help/                   # Subsistema de ajuda, charset MSX->UTF8, DBASEMSG.TXT, MED.HLP e pager
│   ├── msx/                    # Transpilador C/MSXgl, gerador de projetos e conversor IHX -> .COM
│   ├── parser/                 # Analisador léxico/sintático de comandos e cláusulas FOR/WHILE
│   ├── storage/sqlite/         # Driver e persistência de metadados em SQLite
│   ├── ui/
│   │   ├── cli/                # Prompt '.' iterativo, line editor, autocomplete TAB e syntax highlighting
│   │   ├── tui/                # Terminal User Interface em modo texto
│   │   └── gui/                # Interface Desktop Fyne (com stub sob tag nogui)
│   └── version/                # Controle de versão semântica e metadados de build
├── resource/
│   ├── disks/                  # Imagens de disco .DSK originais do dBASE II MSX
│   ├── doc/                    # Manuais escaneados do dBASE II+ MSX e livros de referência em PDF
│   ├── files/                  # DBASEMSG.TXT, MED.HLP, scripts .CMD originais do MSX
│   └── msx/                    # SDK MSXgl e toolchain SDCC embutida
├── bin/                        # Binários executáveis locais gerados (dbase2plus.exe / dbase2plus)
├── dist/                       # Pacotes compactados de distribuição (.zip e .tar.gz)
├── release/                    # Pasta de montagem de arquivos para entrega
├── build.ps1                   # Script de automação e empacotamento Windows (PowerShell)
├── build.sh                    # Script de automação e empacotamento Linux/WSL (Bash)
├── version.json                # Metadados de versão e contador de build
├── SPEC.md                     # Documento vivo de especificação arquitetural
├── README.md                   # Manual de uso e comandos
├── CHANGELOG.md                # Registro histórico de alterações
├── RELEASE.md                  # Notas de release prontas para GitHub Releases
└── LICENSE                     # Licença GPL-3.0
```

---

## ✅ 3. O Que Já Foi Implementado (Status Atual v0.6.4)

1. **Compatibilidade Binária dBASE II (`pkg/dbf`)**:
   - Cabeçalho exato de 521 bytes com assinatura `0x02`.
   - Suporte estrito a tipos `C` (Character), `N` (Numeric) e `L` (Logical).
   - Terminação de registro `0x0D`, delimitação de campos `0x1E` e marcador de deleção `*` / ` `.
   - Operações `PACK`, `ZAP`, `APPEND BLANK`, `RECALL`.
2. **Avaliador de Expressões e Cláusulas Condicionais (`pkg/eval`)**:
   - Operadores aritméticos (`+`, `-`, `*`, `/`), relacionais (`=`, `<>`, `<`, `>`, `<=`, `>=`), lógicos (`.AND.`, `.OR.`, `.NOT.`, `!`) e substring (`$`).
   - Cláusulas `FOR` e `WHILE` integradas aos comandos de dados.
   - Funções intrínsecas: `TRIM`, `UPPER`, `LOWER`, `LEN`, `SUBSTR`, `VAL`, `STR`, `DELETED()`, `RECNO()`, `EOF()`, `DATE()`, `TIME()`, `IIF()`.
   - Atribuição de variáveis via `STORE <expr> TO <var>` ou `<var> = <expr>`.
3. **Criação e Edição Interativa de Dados (`pkg/engine`)**:
   - Fluxo interativo do comando `CREATE` com definição de campos linha a linha (`FIELD NAME,TYPE,WIDTH,DECIMAL PLACES`).
   - Pergunta clássica pós-criação: `INPUT DATA NOW? (Y/N)` com transição direta ao `APPEND`.
   - Telas interativas de inclusão (`APPEND`) e edição (`EDIT [n]`) com suporte a atalhos clássicos (`Ctrl+W`, `Ctrl+Q`, `Ctrl+U`, `Ctrl+E`/`Ctrl+X`, `Ctrl+C`/`Ctrl+R`) e teclas modernas (`Enter`, `Esc`, setas, `Tab`).
4. **Subsistema de Ajuda e Charset MSX (`pkg/help`)**:
   - Conversor de tabela de caracteres da ROM brasileira do MSX para UTF-8.
   - Mais de 86 tópicos do `DBASEMSG.TXT` e do editor `MED.HLP`.
   - Paginador de terminal interativo com navegação por tela e linha.
   - Catálogo estruturado de comandos: `HELP COMMANDS`.
5. **CLI com Autocomplete e Syntax Highlighting (`pkg/ui/cli`)**:
   - Editor de linha de comando com histórico de comandos.
   - Autocomplete cíclico com a tecla `TAB`.
   - Realce de sintaxe colorido ANSI em tempo real (comandos, cláusulas, operadores, funções, strings e erros em vermelho).
6. **Ecossistema MSX, MSXgl e Disquetes Bootáveis (`pkg/msx`)**:
   - Emissor de código C configurado para MSX 2+ (256 KB Memory Mapper) e MSX-DOS 2 via File Handles.
   - Geração de `msxgl_config.h`, `project_config.js` e `Makefile`.
   - Conversor nativo Intel HEX (.ihx) para binário `.COM` em Go puro (`pkg/msx/hex2bin.go`).
   - Gerador autônomo de imagens de disquete de 720 KB FAT12 para MSX-DOS 2 (`pkg/msx/dsk.go`), contendo a distribuição completa de `resource/msxdos2` (`MSXDOS2.SYS`, `COMMAND2.COM`, `MSXDOS.SYS`, subdiretórios `\UTILS` e `\HELP`), `AUTOEXEC.BAT`, `DBASE.COM` (com versão idêntica à do PC), `DBASEMSG.TXT`, tabelas `.DBF` e `README.TXT` pronto para boot no MSX 2+ real.
   - Interface MSX modernizada com caracteres gráficos nativos da ROM do MSX (cantos, linhas simples e duplas, junções) eliminando caracteres legados em ASCII (`+===+` / `+---+`).
   - Rotina de escrita em bloco acelerada via MSX-DOS 2 Function `0x40` (`_WRITE` direcionada ao stdout), eliminando a lentidão das chamadas arcaicas CP/M BDOS 0x09.
   - Subsistema de ajuda completo no MSX com paridade total ao PC/Go: catálogo de comandos e suporte a subtópicos contextuais (`HELP USE`, `HELP BROWSE`, `HELP CREATE`, `HELP LIST`, `HELP EDIT`, etc.).
7. **Automação de Build e Release**:
   - `build.ps1` (Windows / PowerShell) gerando pacotes `.zip` e disquetes `.dsk`.
   - `build.sh` (Linux / Bash) gerando pacotes `.tar.gz` e disquetes `.dsk`.
   - Controle de incremento de versão $Z$ com suporte a `--no-increment` / `-NoIncrement`.

---

## 🗺️ 4. Próximos Passos no Roadmap

Quando iniciar no novo computador, as prioridades recomendadas são:

### 🎯 Passo Imediato Recomendado: Passo 2 — Motor de Scripts `.PRG` e Controle de Fluxo
* **Objetivo**: Permitir executar programas e rotinas completas do dBASE II/MSX no PC e gerar a árvore sintática para compilação C/MSX.
* **Componentes a Implementar**:
  1. Estruturas de controle de fluxo:
     - `DO WHILE <cond>` ... `ENDDO` (com `LOOP` e `EXIT`).
     - `IF <cond>` ... `ELSE` ... `ENDIF`.
  2. Comandos de tela formatada:
     - `@ <linha>,<coluna> SAY <expr>`
     - `@ <linha>,<coluna> GET <var>`
     - `READ`
  3. Execução de arquivos de comando:
     - `DO <arquivo.prg/cmd>` com passagem de parâmetros e pilha de execução.
     - `RETURN` / `CANCEL`.

### 🎯 Passo Seguinte: Passo 3 — Suporte a Índices Binários `.NDX` (B-Tree)
* **Objetivo**: Busca em tempo logarítmico e navegação ordenada de tabelas sem varredura linear.
* **Comandos**: `INDEX ON <chave> TO <arquivo>`, `SET INDEX TO <arquivo>`, `FIND <chave>`, `SEEK <chave>`, `REINDEX`.

---

## 💻 5. Instruções de Configuração no Novo Computador

### 1. Clonar ou Copiar o Repositório
Abra o repositório na pasta desejada no novo PC (ex: `C:\dos\dbase2a` ou diretório de trabalho do Go).

### 2. Abrir no JetBrains GoLand
1. No GoLand, selecione **Open** e aponte para a pasta raiz do projeto.
2. Certifique-se de que o Go SDK (Go 1.22 ou superior) está configurado em **Settings -> Go -> GOROOT**.

### 3. Verificar o Projeto e Executar Testes
No terminal do GoLand (PowerShell ou WSL Bash), execute:
```powershell
# Executar todos os testes unitários
go test -v -tags nogui ./...
```
*Todos os testes devem passar com 100% de sucesso.*

### 4. Compilar e Gerar os Pacotes
* **Windows (PowerShell)**:
  ```powershell
  # Build padrão Windows gerando bin\dbase2plus.exe e dist\*.zip
  .\build.ps1 -NoIncrement
  ```
* **Linux / WSL (Bash)**:
  ```bash
  # Build padrão Linux gerando bin/dbase2plus e dist/*.tar.gz
  ./build.sh --no-increment
  ```

### 5. Executar a Aplicação
```powershell
# Iniciar o modo CLI interativo (prompt .)
.\bin\dbase2plus.exe

# Iniciar o modo Terminal UI
.\bin\dbase2plus.exe --tui

# Executar comando de ajuda direta
.\bin\dbase2plus.exe --exec "HELP COMMANDS"
```

---

## 🤖 6. Instruções para o Junie no Novo Computador

Para retomar o desenvolvimento no novo computador com o Junie:
1. Abra a pasta do projeto no GoLand.
2. Inicie a sessão do Junie e instrua:  
   *"Leia o arquivo `HANDOVER.md` e `SPEC.md` para entender onde paramos. Estamos na versão v0.6.4 'The Exorcist' e vamos iniciar a implementação do Passo 2 (Motor de Scripts .PRG com DO WHILE, IF/ELSE e @ SAY/GET)."*
3. O Junie carregará imediatamente este contexto e continuará sem perda de continuidade.
