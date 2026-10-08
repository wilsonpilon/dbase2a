# dBase II+ — Notas de Release (RELEASE.md)

Este documento registra os pacotes de lançamento, instruções de entrega, metadados de compilação e codinomes temáticos de filmes de terror/horror para o projeto **dBase II+**.

---

## 📽️ Convenção de Codinomes de Horror & Terror

Os lançamentos do projeto são batizados com clássicos do cinema de horror e suspense:

| Versão | Codinome | Ano do Filme | Diretor / Destaque |
|---|---|---|---|
| **v0.1.0** | **Nosferatu** | 1922 | F. W. Murnau — O despertar do dBase II+ |
| **v0.2.0** | **The Shining** (*O Iluminado*) | 1980 | Stanley Kubrick — Expansão da TUI e Engine |
| **v0.3.0** | **Psycho** (*Psicose*) | 1960 | Alfred Hitchcock — GUI Fyne Avançada e Grid |
| **v0.4.0** | **Halloween** (*A Noite do Terror*) | 1978 | John Carpenter — Toolchain MSXgl e Fusion-C |
| **v0.5.0** | **The Thing** (*O Enigma de Outro Mundo*) | 1982 | John Carpenter — Suporte 256KB Memory Mapper |
| **v0.6.0** | **Alien** (*O Oitavo Passageiro*) | 1979 | Ridley Scott — Geração automatizada de .DSK |
| **v0.7.0** | **Suspiria** | 1977 | Dario Argento — Indexador .NDX e Otimizações |
| **v1.0.0** | **The Exorcist** (*O Exorcista*) | 1973 | William Friedkin — Release Estável 1.0 Final |

---

## 📦 Detalhes do Release Atual: `v0.1.0` (*Nosferatu*)

### 🏷️ Metadados do Pacote
- **Versão Semântica**: `v0.1.0`
- **Codinome Oficial**: *Nosferatu*
- **Data de Lançamento**: 08 de Outubro de 2026
- **Licença**: GNU General Public License v3.0 (GPL-3.0)
- **Plataformas Homologadas**: Windows 10/11 (x64), Linux (x86_64 / WSL) e MSX 2+ (MSX-DOS 2 / Z80).

### 🎁 Conteúdo do Pacote de Distribuição (`release/`)
Ao executar o script `build.ps1` (Windows) ou `build.sh` (Linux), a pasta `release/` e o pacote compactado conterão:
1. `dbase2plus.exe` / `dbase2plus` (Binário compilado unificado com suporte a CLI, TUI e GUI).
2. `SPEC.md` (Documento vivo de especificação arquitetural).
3. `README.md` (Manual de uso, comandos e parâmetros).
4. `CHANGELOG.md` (Histórico de alterações).
5. `LICENSE` (Texto da licença GPL-3.0).
6. `version.json` (Metadados de versão e hash de compilação).
7. `resource/` (Templates de código MSX, discos de exemplo e documentação).

### 🚀 Instruções de Execução Rápida
- **Windows**:
  - Extrair o arquivo `dbase2plus-v0.1.0-windows-amd64.zip`.
  - Executar `.\dbase2plus.exe --cli` para abrir o prompt dBase II.
  - Executar `.\dbase2plus.exe --gui` para abrir a interface gráfica Fyne.
- **Linux**:
  - Extrair `tar -xzf dbase2plus-v0.1.0-linux-amd64.tar.gz`.
  - Executar `./dbase2plus --cli` ou `./dbase2plus --gui`.

---
*Para gerar um novo pacote de release, execute `.\build.ps1` no PowerShell ou `./build.sh` no Linux/WSL.*
