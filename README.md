# dBase II++ (PC & MSX Edition)

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)
[![Platform: Windows | Linux | MSX](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20MSX-orange.svg)](#)
[![Go Version: 1.27+](https://img.shields.io/badge/Go-1.27%2B-00ADD8.svg)](#)
[![GUI: Fyne](https://img.shields.io/badge/GUI-Fyne-29BEB0.svg)](#)
[![MSX Target: MSX 2+ / MSX--DOS 2](https://img.shields.io/badge/MSX%20Target-MSX%202%2B%20%2F%20256KB-red.svg)](#)

O **dBase II++** é um clone moderno, poderoso e estendido do clássico **dBase II / dBase II+** (Plus Plus — partindo da versão II+ original e incrementando novas capacidades), projetado como uma ferramenta completa de desenvolvimento, gestão de dados e geração de código tanto para **PC (Windows e Linux)** quanto para o lendário computador **MSX (especialmente MSX 2+ com 256 KB de RAM e MSX-DOS 2)**.

---

### Principais Características

- **Compatibilidade Binária 100% com dBASE II**:
  - Leitura e gravação de arquivos de dados `.DBF` no formato original do dBASE II (cabeçalho 0x02 de 521 bytes, campos de 16 bytes, tipos `C`, `N`, `L`, marcadores de exclusão).
  - Criação interativa de tabelas com comando `CREATE`: definição passo a passo de campos (`001`, `002`...), validação imediata, encerramento com `Enter` vazio e prompt para inserção imediata de dados (`INPUT DATA NOW? (Y/N)`).
  - Inclusão interativa (`APPEND`) e edição (`EDIT`) com tela simples de campos alinhados e encerramento com `Enter` em branco no primeiro campo de novo registro.
  - Totalmente compatível com os arquivos e discos dBase II do MSX.
- **Três Modos de Interface no PC**:
  1. **GUI (Graphical User Interface)**: Construída com **Fyne**, oferecendo grid de dados, edição visual de tabelas, painel de compilação MSX e monitor de banco de dados.
  2. **TUI (Terminal User Interface)**: Modo texto com menus, caixas de diálogo e navegação rápida para terminais modernos no Windows (PowerShell/CMD) e Linux (Bash/Zsh).
  3. **CLI (Command Line Interface)**: O clássico e nostálgico prompt de ponto `.` estendido com:
     - **Autocomplete Cíclico por TAB**: Completa comandos ao digitar prefixos (`cre` -> `CREATE `) e circula entre candidatos a cada pressionamento da tecla TAB (`co` -> `CONTINUE` -> `COPY` -> `COUNT`).
     - **Realce de Sintaxe Dinâmico**: Destaque em cores ANSI em tempo real para comandos, cláusulas, operadores, funções, strings, números e arquivos.
     - **Relatório de Erros em Vermelho**: Mensagens de erro com destaque visual claro (`*** ERRO: ...`).
     - **Paginador Interativo do Help**: Navegação de tópicos longos de ajuda com rolagem por páginas (`Espaço`/`PageDown`/`PageUp`) e linhas (`j`/`k`/setas).
     - **Catálogo de Comandos**: Comando `HELP COMMANDS` com resumo em uma linha de todas as instruções.

---

### Guia de Teclas e Atalhos de Edição

| Tecla Clássica dBASE II | Tecla Moderna / Terminal | Ação Executada |
|---|---|---|
| `<Enter>` | `Enter` | Confirma o campo e avança para o próximo campo / próxima linha da estrutura |
| `<Enter>` (vazio no 1º campo) | `Enter` (em linha vazia) | Finaliza criação de campos no `CREATE` ou encerra inserção no `APPEND` |
| `CTRL-W` ou `CTRL-O` | `Ctrl+W` / `Ctrl+End` / `F10` | Salva o registro atual e sai do modo de edição/append |
| `CTRL-Q` | `Ctrl+Q` / `Esc` | Aborta o registro corrente sem salvar e retorna ao prompt `.` |
| `CTRL-U` | `Ctrl+U` | Alterna a marcação de exclusão lógica (*Deleted* `*` / ` `) |
| `CTRL-X` ou `CTRL-F` | `Seta Baixo` / `Tab` | Avança o cursor para o próximo campo |
| `CTRL-E` ou `CTRL-A` | `Seta Cima` / `Shift+Tab` | Retorna o cursor para o campo anterior |
| `CTRL-C` | `PageDown` | Grava o registro corrente e avança para o próximo registro (modo `EDIT`) |
| `CTRL-R` | `PageUp` | Grava o registro corrente e retorna ao registro anterior (modo `EDIT`) |
- **Persistência de Configurações em SQLite**:
  - Armazenamento de preferências de usuário, histórico de comandos, aliases de tabelas e estatísticas em banco SQLite leve (`dbase2plus.db`).
- **Pipeline de Compilação Cruzada para MSX**:
  - Geração de código fonte em **C (SDCC)** e **Assembly (Z80)** otimizado.
  - Suporte às bibliotecas **Fusion-C (MSX Fusion)** e **MSXgl**.
  - Geração de programas nativos `.COM` para **MSX-DOS 2** tirando proveito dos **256 KB de RAM** do MSX 2+.
- **Automação de Build Multiplataforma**:
  - Script **`build.ps1`** para Windows (PowerShell) com controle automático do contador de compilação ($Z$), criação de pacotes `.zip` e preparação da pasta `release/`.
  - Script **`build.sh`** para Linux / WSL com criação de pacotes `.tar.gz` e incremento de versão.
- **Sistema de Versões Temático**:
  - Esquema $vX.Y.Z$ onde cada lançamento é batizado com o nome de um clássico filme de terror/horror (ex: *Nosferatu*, *The Shining*, *Halloween*, *Psycho*).

---

### Estrutura do Projeto

```
dbase2a/
├── cmd/
│   └── dbase2plus/         # Ponto de entrada principal da aplicação (Main)
├── pkg/
│   ├── config/             # Gerenciamento de configurações e diretórios
│   ├── dbf/                # Leitor/Gravador binário do formato dBASE II (.DBF)
│   ├── engine/             # Interpretador e motor de execução de comandos
│   ├── eval/               # Avaliador polimórfico de expressões matemáticas e lógicas
│   ├── help/               # Sistema de ajuda interativo (DBASEMSG.TXT / MED.HLP) e conversor MSX->UTF8
│   ├── msx/                # Gerador de código e pipeline SDCC/Fusion-C/MSXgl
│   ├── parser/             # Lexer, analisador sintático e AST da linguagem
│   ├── storage/sqlite/     # Camada de banco de dados SQLite (metadados/config)
│   ├── ui/
│   │   ├── cli/            # Interface interativa CLI (prompt .)
│   │   ├── gui/            # Interface gráfica desktop Fyne
│   │   └── tui/            # Interface em modo texto de terminal (TUI)
│   └── version/            # Sistema de controle de versões e apelidos de horror
├── resource/               # Discos .DSK, documentação e utilitários MSX originais
├── scripts/                # Scripts auxiliares
├── build.ps1               # Script de build e empacotamento para Windows
├── build.sh                # Script de build e empacotamento para Linux (WSL)
├── version.json            # Metadados de versão (X.Y.Z, codinome, build)
├── SPEC.md                 # Documento vivo de especificação técnica
├── CHANGELOG.md            # Histórico de alterações e lançamentos
├── RELEASE.md              # Notas e instruções de pacotes de release
└── LICENSE                 # Licença GNU General Public License v3.0
```

---

### Compilação e Instalação

#### No Windows (PowerShell)
Para compilar o projeto, atualizar dependências, incrementar o contador de build ($Z$) e gerar o pacote de distribuição:
```powershell
.\build.ps1
```
Para forçar compilação com um novo codinome ou modo específico:
```powershell
.\build.ps1 -Mode all
```
O executável compilado e o arquivo `.zip` serão gerados na pasta `release/`.

#### No Linux / WSL (Bash)
```bash
chmod +x build.sh
./build.sh
```
O arquivo compactado `.tar.gz` e os binários serão copiados para `release/`.

---

### Parâmetros de Linha de Comando (CLI)

O executável `dbase2plus` suporta múltiplos modos de inicialização:

```bash
# Iniciar o modo CLI interativo (prompt .)
dbase2plus --cli

# Iniciar o modo gráfico Fyne (padrão quando executado com dois cliques)
dbase2plus --gui

# Iniciar o modo Terminal User Interface (TUI)
dbase2plus --tui

# Executar comandos em lote diretamente do terminal
dbase2plus --exec="USE CLIENTES; LIST ALL; QUIT"

# Executar um script dBASE II (.CMD ou .PRG)
dbase2plus --file=FOLHA.CMD

# Gerar código C/MSX a partir de uma tabela DBF
dbase2plus --export-msx=PRODUTOS.DBF --target=msx2p --sdk=fusion-c

# Exibir informações de versão e codinome de terror
dbase2plus --version
```

---

### Licença

Este projeto é software livre e de código aberto sob a licença **GNU General Public License v3.0 (GPL-3.0)**. Consulte o arquivo [LICENSE](LICENSE) para obter o texto integral.
