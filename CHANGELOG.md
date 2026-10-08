# Changelog

Todas as alterações notáveis neste projeto serão documentadas neste arquivo.

O formato é baseado em [Keep a Changelog](https://keepachangelog.com/pt-BR/1.0.0/),
e este projeto adota versionamento no padrão `vX.Y.Z` onde:
- `X` = Pacote de funcionalidades maiores / fases arquiteturais
- `Y` = Novo recurso (feature) introduzido ou finalizado
- `Z` = Contador de compilação/build incrementado automaticamente a cada build
- Cada lançamento recebe o codinome de um clássico filme de terror/horror.

---

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
