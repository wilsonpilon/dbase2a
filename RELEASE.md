# dBase II++ — Notas de Release (RELEASE.md)

Este documento registra os pacotes de lançamento, instruções de entrega, metadados de compilação e codinomes temáticos de filmes de terror/horror para o projeto **dBase II++**.

---

## 📽️ Convenção de Codinomes de Horror & Terror

Os lançamentos do projeto são batizados com clássicos do cinema de horror e suspense:

| Versão | Codinome | Ano do Filme | Diretor / Destaque |
|---|---|---|---|
| **v0.1.0** | **Nosferatu** | 1922 | F. W. Murnau — O despertar do dBase II++ |
| **v0.2.0** | **The Shining** (*O Iluminado*) | 1980 | Stanley Kubrick — Avaliador de Expressões e FOR/WHILE |
| **v0.3.0** | **Psycho** (*Psicose*) | 1960 | Alfred Hitchcock — MSXgl e SDCC 4.6.0 |
| **v0.4.0** | **Halloween** (*A Noite do Terror*) | 1978 | John Carpenter — Ajuda Interativa e Charset MSX |
| **v0.5.0** | **The Thing** (*O Enigma de Outro Mundo*) | 1982 | John Carpenter — Paginador, Autocomplete e Syntax |
| **v0.6.0** | **The Exorcist** (*O Exorcista*) | 1973 | William Friedkin — CREATE interativo, APPEND/EDIT e dBase II++ |
| **v0.7.0** | **Alien** (*O Oitavo Passageiro*) | 1979 | Ridley Scott — Geração automatizada de .DSK |
| **v1.0.0** | **Suspiria** | 1977 | Dario Argento — Release Estável 1.0 Final |

---

## 🚀 GitHub Release — Texto para Publicação

> **Copie o conteúdo abaixo diretamente na descrição do novo Release no GitHub:**

```markdown
# dBase II++ v0.6.4 "The Exorcist" — Release Oficial

Uma reimplementação moderna, rápida e estendida do clássico **dBase II / dBase II+** para **PC (Windows e Linux)** e ecossistema **MSX (MSX 2+ com 256 KB RAM e MSX-DOS 2)**, escrita em Go.

---

### 🌟 Principais Recursos & Novidades

#### 1. Criação e Edição Interativa de Dados (`CREATE`, `APPEND`, `EDIT`)
* **Fluxo Clássico `CREATE`**:
  * Entrada sequencial e assistida de campos (`FIELD NAME,TYPE,WIDTH,DECIMAL PLACES`).
  * Validação em tempo real de tipos (`C`, `N`, `L`), tamanhos e casas decimais.
  * Pergunta clássica pós-criação: `INPUT DATA NOW? (Y/N)` com transição instantânea para inclusão.
* **Telas de Dados `APPEND` & `EDIT`**:
  * Entrada campo a campo com indicação de tamanhos e valores preexistentes.
  * Finalização rápida ao pressionar `Enter` no primeiro campo em branco.
  * Mapeamento de atalhos clássicos do dBASE II (`Ctrl+W` para salvar/sair, `Ctrl+Q` para cancelar, `Ctrl+U` para exclusão lógica) e teclas de terminal modernas (`Enter`, `Esc`, setas, `Tab`).

#### 2. Subsistema de Ajuda Interativo com Charset MSX e Paginador (`HELP`)
* Decodificação transparente da ROM de caracteres de 8-bits do MSX brasileiro (Gradiente/Sharp) para UTF-8 no PC, com acentuação e cedilhas perfeitas.
* Mais de 86 tópicos indexados de `DBASEMSG.TXT` e do editor `MED.HLP`.
* Paginador interativo no terminal com rolagem por página (`Espaço`, `PageDown`/`PageUp`) e linha (`Setas`, `j`/`k`).
* Novo catálogo `HELP COMMANDS` com resumo de uma linha para cada comando.

#### 3. CLI Moderna com Autocomplete e Syntax Highlighting
* Autocomplete cíclico inteligente via tecla `TAB` para comandos dBase.
* Realce de sintaxe colorido ANSI em tempo real: Comandos (Ciano), Cláusulas (Branco Negrito), Operadores (Amarelo), Funções (Magenta), Arquivos/Strings (Verde), Números (Azul) e Mensagens de Erro em Vermelho Brilhante.

#### 4. Avaliador de Expressões Polimórfico (`pkg/eval`)
* Suporte a operadores lógicos (`.AND.`, `.OR.`, `.NOT.`), relacionais (`=`, `<>`, `<`, `>`, `<=`, `>=`), aritméticos (`+`, `-`, `*`, `/`) e substring (`$`).
* Cláusulas condicionais `FOR` e `WHILE` em `LIST`, `DISPLAY`, `COUNT`, `SUM`, `LOCATE`, `REPLACE`, `DELETE` e `RECALL`.
* Funções intrínsecas: `TRIM`, `UPPER`, `LOWER`, `LEN`, `SUBSTR`, `VAL`, `STR`, `DELETED()`, `RECNO()`, `EOF()`, `DATE()`, `TIME()`, `IIF()`.

#### 5. Pipeline MSXgl e Compilação MSX 2+
* Transpilador para código C moderno otimizado para o compilador SDCC 4.6.0 e framework **MSXgl**.
* Suporte nativo ao **Memory Mapper de 256 KB** do MSX 2+ e **MSX-DOS 2** via File Handles.
* Conversor Intel HEX (.ihx) para binário `.COM` embutido em Go puro.

---

### 📦 Arquivos do Release

| Arquivo | Plataforma | Descrição |
|---|---|---|
| `dbase2plus-v0.6.4-windows-amd64.zip` | Windows (x64) | Executável Windows (`dbase2plus.exe`), disquete MSX `.dsk`, documentação e recursos |
| `dbase2plus-v0.6.4-linux-amd64.tar.gz` | Linux (x86_64) | Executável Linux (`dbase2plus`), disquete MSX `.dsk`, documentação e recursos |
| `dbase2plus-v0.6.4-msx2p-dos2.dsk` | MSX 2+ (720 KB FAT12) | Imagem de disquete de boot com MSX-DOS 2 (`MSXDOS2.SYS`, `COMMAND2.COM`), utilitários (`\UTILS`), manuais (`\HELP`), `DBASE.COM`, `AUTOEXEC.BAT` e dados de teste |

---

### ⚡ Como Executar

#### No MSX Real (ou openMSX / BlueMSX):
1. Grave a imagem `dbase2plus-v0.6.4-msx2p-dos2.dsk` (ou `release/dbase2plus.dsk`) em um disquete de 3.5" de 720KB (ou carregue no emulador no drive A:).
2. Ligue o MSX 2+ com o disquete inserido. O MSX-DOS 2 iniciará e executará automaticamente o `AUTOEXEC.BAT`, configurando modo texto de 80 colunas e abrindo o prompt `.` do dBase II++.
3. No prompt, digite `USE ARQUIVO`, `LIST`, `BROWSE` ou `HELP`.

#### No Windows:
```cmd
# Extrair o arquivo dbase2plus-v0.6.4-windows-amd64.zip
.\dbase2plus.exe          # Inicia no modo padrão (CLI)
.\dbase2plus.exe --tui    # Inicia no modo Terminal UI
.\dbase2plus.exe --gui    # Inicia no modo Interface Gráfica Fyne
.\dbase2plus.exe --make-dsk # Gera disquete bootável MSX-DOS 2 de 720KB
.\dbase2plus.exe --exec "HELP" # Executa comando direto
```

#### Linux:
```bash
# Extrair o arquivo dbase2plus-v0.6.4-linux-amd64.tar.gz
tar -xzf dbase2plus-v0.6.4-linux-amd64.tar.gz
./dbase2plus --cli
```
```

---

## 📦 Detalhes do Release Atual: `v0.6.4` (*The Exorcist*)

### 🏷️ Metadados do Pacote
- **Versão Semântica**: `v0.6.4`
- **Codinome Oficial**: *The Exorcist*
- **Data de Lançamento**: 08 de Outubro de 2026
- **Licença**: GNU General Public License v3.0 (GPL-3.0)
- **Plataformas Homologadas**: Windows 10/11 (x64), Linux (x86_64 / WSL) e MSX 2+ (MSX-DOS 2 / Z80).

### 🎁 Conteúdo do Pacote de Distribuição (`dist/` e `release/`)
1. `dbase2plus.exe` / `dbase2plus` (Binário compilado unificado com suporte a CLI, TUI e GUI).
2. `SPEC.md` (Documento vivo de especificação arquitetural).
3. `README.md` (Manual de uso, comandos e parâmetros).
4. `CHANGELOG.md` (Histórico de alterações).
5. `RELEASE.md` (Notas de lançamento).
6. `HANDOVER.md` (Instruções completas para transição de máquina).
7. `LICENSE` (Texto da licença GPL-3.0).
8. `version.json` (Metadados de versão e hash de compilação).
9. `resource/` (Templates de código MSX, discos de exemplo e documentação).

---
*Para gerar novos pacotes de release, utilize `.\build.ps1 -NoIncrement` no PowerShell ou `./build.sh --no-increment` no Linux/WSL.*
