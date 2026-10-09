# ESPECIFICAÇÃO TÉCNICA E ARQUITETURAL (SPEC.md)
## Projeto dBase II++ (PC & MSX 2+)

> **Documento Vivo (Live Document)**  
> Este documento é atualizado continuamente ao longo do ciclo de vida do projeto para registrar arquitetura, especificações técnicas, decisões de engenharia, progresso de implementação e pendências futuras.

---

### 1. Visão Geral e Objetivos do Projeto

O **dBase II++** é uma reimplementação moderna e estendida do clássico sistema gerenciador de banco de dados relacional e linguagem de programação **dBase II / dBase II+** (Plus Plus — partindo do dBase II+ original e incrementando novos recursos e tecnologias), com foco duplo:
1. **Ambiente PC (Windows & Linux)**:
   - Aplicação de alto desempenho escrita em **Go (Golang)**.
   - Suporte a múltiplos modos de interface: **GUI Desktop** (utilizando o framework **Fyne**), **TUI** (Terminal User Interface completa para consoles de texto) e **CLI Interativo** (com o clássico prompt de ponto `.` e processador de linha de comando extensivo).
   - Persistência e gerenciamento de configurações, metadados, catálogo de tabelas e logs através de **SQLite**.
   - Leitura, edição, gravação e manipulação nativa e 100% compatível com arquivos binários de dados `.DBF` e comandos no padrão dBASE II original.
2. **Ecossistema MSX (MSX 1, MSX 2, MSX 2+ e MSX Turbo R)**:
   - Compatibilidade total com o dBase II original do MSX (incluindo discos de sistema, overlays e arquivos de dados).
   - Geração de novos clientes, frontends e executáveis `.COM` modernos para **MSX-DOS 2**, otimizados para **MSX 2+** com **256 KB de RAM** mapeada (Memory Mapper).
   - Pipeline de compilação cruzada no PC integrando **SDCC (Small Device C Compiler)**, **MSX Fusion (Fusion-C)** e **MSXgl**, permitindo escrever rotinas e programas no PC e compilar binários nativos Z80 otimizados para o MSX.

---

### 2. Arquitetura do Sistema

```
+-----------------------------------------------------------------------------------+
|                                 dBase II++ Engine (Go)                            |
+-----------------------------------------------------------------------------------+
|  +-------------------+  +-------------------+  +-------------------------------+  |
|  |     GUI (Fyne)    |  |     TUI (Term)    |  |       CLI (Dot Prompt .)      |  |
|  +-------------------+  +-------------------+  +-------------------------------+  |
|            |                      |                            |                  |
|            +----------------------+----------------------------+                  |
|                                   v                                               |
|                    +-----------------------------+                                |
|                    |     CLI Flag / Mode Router  |                                |
|                    +-----------------------------+                                |
|                                   |                                               |
|       +---------------------------+---------------------------+                   |
|       v                           v                           v                   |
|  +------------------+    +------------------+    +-----------------------+        |
|  | Parser & Lexer   |    | Execution Engine |    | MSX Toolchain Bridge  |        |
|  | (dBASE II Syntax)|    | (Commands/States)|    | (SDCC/MSXgl/Fusion-C) |        |
|  +------------------+    +------------------+    +-----------------------+        |
|       |                           |                           |                   |
|       v                           v                           v                   |
|  +------------------+    +------------------+    +-----------------------+        |
|  |  DBF Engine v2   |    | SQLite Metadata  |    | MSX-DOS 2 .COM Target |        |
|  | (Raw Binary I/O) |    | & Config Store   |    | (MSX 2+ 256KB Mapper) |        |
|  +------------------+    +------------------+    +-----------------------+        |
+-------|---------------------------|---------------------------|-------------------+
        v                           v                           v
 [ Arquivos .DBF ]         [ dbase2plus.db ]          [ Discos .DSK / .COM ]
```

---

### 3. Especificação do Formato de Arquivo dBASE II (.DBF)

Para garantir **100% de compatibilidade binária** com os discos e arquivos dBase II do MSX e CP/M:

#### 3.1. Estrutura do Cabeçalho (Header)
O cabeçalho do dBASE II possui tamanho fixo de **521 bytes**:
- **Byte 0**: Versão do formato (`0x02` para dBASE II).
- **Bytes 1-2**: Contagem total de registros na tabela (`uint16`, Little-Endian, máximo de 65.535 registros).
- **Bytes 3-5**: Data da última atualização no formato `YY MM DD` (1 byte para ano, 1 byte para mês, 1 byte para dia).
- **Bytes 6-7**: Tamanho do registro de dados em bytes (`uint16`, Little-Endian, incluindo o byte do indicador de deleção).
- **Bytes 8-519**: Descritores de campos (até **32 campos** no máximo). Cada descritor ocupa **16 bytes**:
  - `Bytes 0-10`: Nome do campo (até 10 caracteres ASCII, preenchido com `0x00`).
  - `Byte 11`: Tipo de dado ASCII (`'C'` = Caractere, `'N'` = Numérico, `'L'` = Lógico).
  - `Byte 12`: Tamanho do campo em bytes (`1` a `254`).
  - `Bytes 13-14`: Endereço de memória em tempo de execução ou deslocamento de dados (Little-Endian).
  - `Byte 15`: Contagem de casas decimais para campos numéricos (`0` a `19`).
- **Byte 520**: Terminador de cabeçalho (`0x0D` / Carriage Return).

#### 3.2. Estrutura dos Registros de Dados
- Cada registro inicia com 1 byte de marcador de status:
  - `0x20` (Espaço em branco ASCII): Registro ativo (não deletado).
  - `0x2A` (Asterisco `*` ASCII): Registro marcado como excluído (deletado logicamente; removido fisicamente apenas após `PACK`).
- Os campos sucedem sequencialmente em formato ASCII puro, preenchidos com espaços à direita para texto (`C`) e espaços à esquerda para números (`N`).
- O final do arquivo é demarcado pelo caractere End-Of-File do CP/M/MSX-DOS (`0x1A` / `EOF` / `Ctrl+Z`).

---

### 4. Engine de Execução e Gramática dBASE II

O motor de comandos implementa a sintaxe padrão do dBASE II com suporte interativo e em lote (`.CMD` / `.PRG`):

| Comando | Descrição | Status no PC | Status no MSX Generator |
|---|---|---|---|
| `CREATE` | Cria uma nova estrutura de tabela `.DBF` | Concluído | Concluído |
| `USE [arq]` | Abre uma tabela de dados (com ou sem índice) | Concluído | Concluído |
| `LIST [campos] [FOR/WHILE]` | Lista registros na tela/impressora com filtros | Concluído | Concluído |
| `DISPLAY [RECNO/ALL]` | Exibe registro atual ou conjunto de registros | Concluído | Concluído |
| `APPEND [BLANK]` | Insere novos registros de dados | Concluído | Concluído |
| `EDIT [recno]` | Edita os valores do registro atual ou informado | Concluído | Concluído |
| `BROWSE` | Modo de navegação tabular em tela cheia | Concluído | Concluído |
| `DELETE [scope]` | Marca registros como excluídos (`*`) | Concluído | Concluído |
| `RECALL [scope]` | Restaura registros marcados para exclusão | Concluído | Concluído |
| `PACK` | Remove fisicamente os registros excluídos e recompacta | Concluído | Concluído |
| `ZAP` | Esvazia todos os registros da tabela ativa | Concluído | Concluído |
| `REPLACE [campo] WITH [val]` | Atualiza campos com base em expressões | Concluído | Concluído |
| `INDEX ON [expr] TO [ndx]` | Gera ou atualiza índice de pesquisa rápida | Concluído | Planejado |
| `FIND / SEEK [chave]` | Pesquisa rápida via índice | Concluído | Planejado |
| `LOCATE FOR [cond]` | Pesquisa sequencial baseada em condição | Concluído | Concluído |
| `CONTINUE` | Continua a busca do último `LOCATE` | Concluído | Concluído |
| `COUNT [scope] [FOR cond]` | Conta ocorrências de registros | Concluído | Concluído |
| `SUM [campos] [FOR cond]` | Totaliza colunas numéricas | Concluído | Concluído |
| `DO [programa.cmd]` | Executa script de comandos dBASE II | Concluído | Concluído |
| `SET [param] [ON/OFF]` | Configura parâmetros de ambiente (TALK, ECHO, etc.) | Concluído | Concluído |
| `QUIT` | Salva todos os buffers e encerra a aplicação | Concluído | Concluído |

---

### 5. Modos de Operação no PC

A aplicação no PC disponibiliza três modos de interface integrados:

1. **Modo CLI (Command Line Interface)**:
   - Emulação fiel do terminal dBase II com prompt de comando `.`.
   - Editor de linha interativo (`LineEditor`) com suporte a histórico, atalhos de navegação e edição.
   - **Autocomplete Cíclico por TAB**: Autocompleta comandos com único match (`cre` -> `CREATE `) ou circula entre múltiplos candidatos a cada TAB sucessivo (`co` -> `CONTINUE` -> `COPY` -> `COUNT`...). Suporte contextual para comandos, tópicos de help, parâmetros de `SET` e nomes de arquivos.
   - **Realce de Sintaxe Dinâmico (Syntax Highlighting)**: Destaque de comandos (Ciano), cláusulas (Branco Negrito), operadores (Amarelo), funções (Magenta), arquivos/strings (Verde), números (Azul) e mensagens de erro destacadas em Vermelho Brilhante.
   - Suporte a saída paginada interativa (`pkg/help/pager.go`) com rolagem por páginas e linhas.
   - Execução direta de parâmetros passados via terminal (`--cli`, `--exec="USE CLIENTES; LIST"`, `--file=rotina.cmd`).
2. **Modo TUI (Terminal User Interface)**:
   - Interface em modo texto formatada com caixas, menus suspensos, status bar e visualizador de tabelas para terminais ANSI/VT100 no Windows PowerShell e Linux Bash.
3. **Modo GUI (Graphical User Interface em Fyne)**:
   - Interface gráfica moderna, multiplataforma (Windows e Linux).
   - Grid de dados dinâmico para navegação e edição de arquivos `.DBF`.
   - Painel de gerenciamento de compilação MSX (seleção de target, memórias, SDKs SDCC/MSXgl/Fusion-C, visualizador de ASM/C e compilação para `.COM` e `.DSK`).
   - Console interativo integrado com realce de sintaxe.

---

### 6. Pipeline de Compilação Cruzada para MSX

O subsistema MSX permite gerar programas para a arquitetura Z80 / MSX-DOS:
- **Alvo Principal**: MSX 2+ com 256 KB RAM (Memory Mapper de 16 KB por segmento) e MSX-DOS 2.
- **SDKs Suportados**:
  - **Fusion-C (MSX Fusion)**: Bibliotecas de acesso direto ao VDP (modos de tela 0 a 8), bios, PSG e disco.
  - **MSXgl**: Biblioteca ultra-otimizada para jogos e aplicações de alta performance em C/ASM.
  - **SDCC (Small Device C Compiler)**: Geração de código de máquina Z80 nativo empacotado em executáveis binários `.COM`.
- **Modos de Geração**:
  1. *Direto dBASE II*: Gera scripts de comando compatíveis com o interpretador clássico `DBASE.COM` do MSX.
  2. *Executável Nativo Otimizado (.COM)*: Compila as tabelas de dados, queries e rotinas de interface para código de máquina C/ASM, utilizando o Memory Mapper para carregar tabelas inteiras em memória RAM expandida de 256 KB.

---

### 7. Sistema de Versionamento e Releases

O versionamento segue a regra estrita definida para o projeto:
$$\text{Formato: } \mathbf{vX.Y.Z}$$
- **$X$ (Major)**: Pacote de funcionalidades maiores ou mudança de fase da arquitetura.
- **$Y$ (Minor)**: Novo recurso/feature incluído ou finalizado.
- **$Z$ (Patch / Build)**: Incrementado automaticamente a cada ciclo de compilação/build pelos scripts `build.ps1` e `build.sh`.
- **Codinomes**: Cada versão lançada recebe como apelido o título de um clássico filme de terror/horror (ex: *Nosferatu*, *The Shining*, *Psycho*, *Halloween*, *Alien*, *The Thing*, *Suspiria*, *Exorcist*).

---

### 8. Divisão do Projeto e Status de Implementação (Roadmap)

#### Fase 1: Fundação, Arquitetura e Formato DBF (Concluído / Atual)
- [x] Estrutura modular de diretórios em Go (`cmd/`, `pkg/dbf/`, `pkg/engine/`, `pkg/storage/`, `pkg/ui/`, `pkg/msx/`, `pkg/version/`).
- [x] Licenciamento em GPL v3.0 (`LICENSE`).
- [x] Documentação inicial completa (`SPEC.md`, `README.md`, `CHANGELOG.md`, `RELEASE.md`).
- [x] Parser e gerador de arquivos binários dBASE II `.DBF` compatível com MSX (cabeçalho 0x02, campos C/N/L, registros e marcadores).
- [x] Camada de persistência e configuração em SQLite (`dbase2plus.db`).
- [x] Módulo de controle de versão semântica e apelidos de filmes de terror (`version.json` e `pkg/version`).
- [x] Scripts de automação de build para Windows (`build.ps1`) e Linux (`build.sh`) com incremento automático de $Z$ e empacotamento ZIP/TAR.GZ.
- [x] Interface CLI com prompt `.` e interpretador de comandos básico (`CREATE`, `USE`, `LIST`, `APPEND`, `QUIT`).
- [x] Módulo gerador de código C/ASM para MSX (Fusion-C / MSXgl).

#### Fase 2: Expansão da Engine, TUI e GUI Fyne (Concluído / Atual)
- [x] Implementação dos comandos estendidos (`PACK`, `ZAP`, `REPLACE`, `LOCATE`, `COUNT`, `SUM`, `DO`, `STORE`, `?`, `??`, `GOTO`, `SKIP`).
- [x] Interface TUI avançada para terminal via ANSI/VT100.
- [x] Interface gráfica Desktop via **Fyne** com tabela de dados e editor de consultas.
- [x] Avaliador de expressões lógicas e matemáticas completas (`.AND.`, `.OR.`, `.NOT.`, `>`, `<`, `=`, `<=`, `>=`, `<>`, `#`, `$`, `+`, `-`, `*`, `/`, funções intrínsecas e filtros `FOR`/`WHILE`).
- [x] Subsistema de Ajuda Dinâmica (`pkg/help`) com suporte total a `DBASEMSG.TXT`, `MED.HLP` e conversor transparente da ROM de caracteres de 8-bits do MSX nacional para UTF-8.
- [x] Paginador Interativo do Help com navegação por teclado (`pkg/help/pager.go`) e catálogo `HELP COMMANDS`.
- [x] Editor CLI interativo com Autocomplete Cíclico por TAB e Realce de Sintaxe Dinâmico / Erros em Vermelho (`pkg/ui/cli`).
- [x] Criação Interativa de Tabelas (`CREATE`) com cabeçalho de estrutura, validação de campos passo a passo (`001`, `002`), término com `Enter` vazio e prompt de inclusão imediata (`INPUT DATA NOW? (Y/N)`).
- [x] Modo Interativo de Inclusão (`APPEND`) e Edição (`EDIT`) com tela de campos alinhados, encerramento por `Enter` no 1º campo e mapeamento completo de teclas clássicas (Ctrl) e modernas.
- [ ] Suporte a arquivos de índices `.NDX` do dBASE II original (B-Tree).
- [ ] Interpretador de scripts e estruturas de controle de fluxo `.PRG` (`DO WHILE ... ENDDO`, `IF ... ELSE ... ENDIF`).

#### Fase 3: Toolchain MSX Avançada e Mapeamento de Memória 256KB
- [x] Integração completa e suporte padrão à SDK **MSXgl** moderna com suporte a SDCC 4.6.0.
- [x] Geração de templates de código C para MSX-DOS 2 com suporte a telas texto (Screen 0) de 80 colunas.
- [x] Mapeamento dinâmico de segmentos de 16 KB no Memory Mapper do MSX 2+ (`dos_mapper.h`, `DOSMapper_Alloc`, `DOSMapper_SetPage2`) para cache de tabelas grandes em 256 KB RAM.
- [x] Conversor nativo em Go de Intel HEX (.ihx) para binário executável MSX-DOS 2 (.COM) sem dependências externas (`pkg/msx/hex2bin.go`).
- [x] Geração de `msxgl_config.h`, `project_config.js`, `Makefile`, `build.bat` e `build.sh` para 1-click build no PC e MSX.
- [x] Criação autônoma de imagens de disco virtuais `.DSK` (720 KB FAT12 / MSX-DOS 2) contendo `MSXDOS2.SYS`, `AUTOEXEC.BAT`, `DBASE.COM`, `DBASEMSG.TXT`, tabelas `.DBF` e `README.TXT` para boot no MSX real e emuladores (`pkg/msx/dsk.go` e flags `--dsk`/`--make-dsk`).
- [ ] Integração de emulador MSX (openMSX / BlueMSX) diretamente disparado pela GUI do Fyne para testes instantâneos.

---
*Última atualização: 2026-10-08 | Versão: v0.6.4 "The Exorcist" | Consulte também: HANDOVER.md*
