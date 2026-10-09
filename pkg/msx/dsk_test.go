package msx

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"dbase2a/pkg/version"
)

func TestFloppyDisk_CreationAndGeometry(t *testing.T) {
	disk := NewFloppyDisk("TESTEDSK")

	if len(disk.Image) != DiskSizeBytes {
		t.Fatalf("Tamanho esperado da imagem de disco: %d bytes, obtido %d", DiskSizeBytes, len(disk.Image))
	}

	// 1. Validação do Boot Sector (Setor 0)
	bootSector := disk.Image[:512]

	// Jump instrução
	if bootSector[0] != 0xEB || bootSector[1] != 0xFE || bootSector[2] != 0x90 {
		t.Errorf("Instrução Jump incorreta no setor de boot: %02X %02X %02X", bootSector[0], bootSector[1], bootSector[2])
	}

	// OEM ID
	oem := string(bootSector[3:11])
	if oem != "MSXDOS2 " {
		t.Errorf("OEM ID esperado 'MSXDOS2 ', obtido '%s'", oem)
	}

	// Bytes por setor
	bps := binary.LittleEndian.Uint16(bootSector[11:13])
	if bps != 512 {
		t.Errorf("Bytes por setor esperados: 512, obtido: %d", bps)
	}

	// Setores por cluster
	spc := bootSector[13]
	if spc != 2 {
		t.Errorf("Setores por cluster esperados: 2, obtido: %d", spc)
	}

	// Total setores
	totSec := binary.LittleEndian.Uint16(bootSector[19:21])
	if totSec != 1440 {
		t.Errorf("Total de setores esperado: 1440, obtido: %d", totSec)
	}

	// Media Descriptor
	media := bootSector[21]
	if media != 0xF9 {
		t.Errorf("Media descriptor esperado: 0xF9, obtido: 0x%02X", media)
	}

	// 2. Validação da FAT1 e FAT2
	fat1Offset := ReservedSectors * SectorSize
	if disk.Image[fat1Offset] != 0xF9 || disk.Image[fat1Offset+1] != 0xFF || disk.Image[fat1Offset+2] != 0xFF {
		t.Errorf("Cabeçalho da FAT1 inválido: %02X %02X %02X",
			disk.Image[fat1Offset], disk.Image[fat1Offset+1], disk.Image[fat1Offset+2])
	}

	fat2Offset := (ReservedSectors + SectorsPerFAT) * SectorSize
	if disk.Image[fat2Offset] != 0xF9 || disk.Image[fat2Offset+1] != 0xFF || disk.Image[fat2Offset+2] != 0xFF {
		t.Errorf("Cabeçalho da FAT2 inválido: %02X %02X %02X",
			disk.Image[fat2Offset], disk.Image[fat2Offset+1], disk.Image[fat2Offset+2])
	}
}

func TestFloppyDisk_AddFileAndReadBack(t *testing.T) {
	disk := NewFloppyDisk("TESTEDSK")

	content := []byte("ESTE E UM TESTE DE ARQUIVO NO DISCO MSX-DOS 2 DE 720KB\r\n")
	err := disk.AddFile("TESTE.TXT", content)
	if err != nil {
		t.Fatalf("Erro ao adicionar arquivo ao disquete: %v", err)
	}

	// Verifica se a entrada está no diretório raiz
	rootDirOffset := (ReservedSectors + NumberOfFATs*SectorsPerFAT) * SectorSize
	rootDirBytes := disk.Image[rootDirOffset : rootDirOffset+RootDirSectors*SectorSize]

	// Entrada 1 (depois do volume label na entrada 0)
	entry := rootDirBytes[32:64]
	name := string(entry[0:11])
	if name != "TESTE   TXT" {
		t.Errorf("Nome 8.3 esperado 'TESTE   TXT', obtido '%s'", name)
	}

	fileSize := binary.LittleEndian.Uint32(entry[28:32])
	if int(fileSize) != len(content) {
		t.Errorf("Tamanho do arquivo esperado: %d, obtido: %d", len(content), fileSize)
	}

	firstCluster := binary.LittleEndian.Uint16(entry[26:28])
	if firstCluster != 2 {
		t.Errorf("Primeiro cluster esperado: 2, obtido: %d", firstCluster)
	}

	// Verifica os dados gravados no cluster 2
	dataAreaOffset := (ReservedSectors + NumberOfFATs*SectorsPerFAT + RootDirSectors) * SectorSize
	clusterData := disk.Image[dataAreaOffset : dataAreaOffset+len(content)]
	if !bytes.Equal(clusterData, content) {
		t.Errorf("Conteúdo do arquivo no disco não confere com os dados gravados")
	}
}

func TestFloppyDisk_Subdirectories(t *testing.T) {
	disk := NewFloppyDisk("SUBDIRTEST")

	subContent := []byte("CONTEUDO DENTRO DO SUBDIRETORIO UTILS\r\n")
	err := disk.AddPath("UTILS/CHKDSK.COM", subContent, AttrArchive)
	if err != nil {
		t.Fatalf("Erro ao adicionar arquivo em subdiretório: %v", err)
	}

	// Verifica se o diretório UTILS foi criado no diretório raiz
	rootDirOffset := (ReservedSectors + NumberOfFATs*SectorsPerFAT) * SectorSize
	rootDirBytes := disk.Image[rootDirOffset : rootDirOffset+RootDirSectors*SectorSize]

	entry1 := rootDirBytes[32:64]
	dirName := string(entry1[0:11])
	if dirName != "UTILS      " {
		t.Errorf("Nome do diretório esperado 'UTILS      ', obtido '%s'", dirName)
	}
	if entry1[11]&AttrDirectory == 0 {
		t.Errorf("Atributo de diretório não encontrado para UTILS")
	}
}

func TestBuildMSXDOS2BootDisk(t *testing.T) {
	tempDir := t.TempDir()
	dskPath := filepath.Join(tempDir, "msxdos2_boot.dsk")

	customFiles := map[string][]byte{
		"EXTRA.TXT": []byte("Arquivo adicional de teste\r\n"),
	}

	err := BuildMSXDOS2BootDisk(dskPath, customFiles)
	if err != nil {
		t.Fatalf("Erro ao construir disco bootável MSX-DOS 2: %v", err)
	}

	info, err := os.Stat(dskPath)
	if err != nil {
		t.Fatalf("Erro ao verificar arquivo DSK gerado: %v", err)
	}

	if info.Size() != DiskSizeBytes {
		t.Errorf("Tamanho do arquivo DSK gerado deve ser exatamente %d bytes (720 KB), obtido %d",
			DiskSizeBytes, info.Size())
	}

	// Verifica se VRAM.ALF está presente no disco
	dskBytes, err := os.ReadFile(dskPath)
	if err != nil {
		t.Fatalf("Erro ao ler DSK: %v", err)
	}
	rootDirOffset := (ReservedSectors + NumberOfFATs*SectorsPerFAT) * SectorSize
	rootDirBytes := dskBytes[rootDirOffset : rootDirOffset+RootDirSectors*SectorSize]
	foundVRAM := false
	for i := 0; i < RootDirectoryEntries; i++ {
		entry := rootDirBytes[i*32 : (i+1)*32]
		if string(entry[0:11]) == "VRAM    ALF" {
			foundVRAM = true
			break
		}
	}
	if !foundVRAM {
		t.Errorf("Arquivo VRAM.ALF não encontrado no diretório raiz do disquete")
	}
}

func TestGenerateMSXExecutable(t *testing.T) {
	v := version.Info{
		Major:     0,
		Minor:     6,
		Patch:     4,
		Codename:  "The Exorcist",
		BuildDate: "2026-10-08",
		Commit:    "9b9f18d",
	}

	bin := GenerateMSXExecutable(v)
	if len(bin) == 0 {
		t.Fatalf("Binário MSX gerado está vazio")
	}

	if bin[0] != 0xC3 {
		t.Errorf("Primeiro byte do executável MSX-DOS .COM deve ser JP (0xC3), obtido 0x%02X", bin[0])
	}

	// Verifica se a versão e o codinome estão no binário
	binStr := string(bin)
	if !bytes.Contains(bin, []byte("v0.6.4")) {
		t.Errorf("Binário MSX deve conter a versão 'v0.6.4'")
	}
	if !bytes.Contains(bin, []byte("The Exorcist")) {
		t.Errorf("Binário MSX deve conter o codinome 'The Exorcist'")
	}
	if !bytes.Contains(bin, []byte("256KB")) {
		t.Errorf("Binário MSX deve referenciar 256KB Memory Mapper")
	}

	// Verifica se o prompt é ". $" sem quebra de linha imediata
	if !bytes.Contains(bin, []byte(". $")) {
		t.Errorf("Binário MSX deve conter prompt '. $'")
	}

	// 1. Verifica presença de caracteres gráficos da ROM do MSX (caixas e molduras nativas)
	if !bytes.Contains(bin, []byte{0xDA}) || !bytes.Contains(bin, []byte{0xC4}) || !bytes.Contains(bin, []byte{0xBF}) {
		t.Errorf("Binário MSX deve conter caracteres gráficos de moldura simples da ROM MSX (0xDA, 0xC4, 0xBF)")
	}
	if !bytes.Contains(bin, []byte{0xC9}) || !bytes.Contains(bin, []byte{0xCD}) || !bytes.Contains(bin, []byte{0xBB}) {
		t.Errorf("Binário MSX deve conter caracteres gráficos de moldura dupla da ROM MSX (0xC9, 0xCD, 0xBB)")
	}
	if bytes.Contains(bin, []byte("+------+")) || bytes.Contains(bin, []byte("+===+")) {
		t.Errorf("Binário MSX não deve conter molduras em ASCII arcaico '+---+' ou '+===+'")
	}

	// 2. Verifica presença do catálogo completo de comandos no HELP (igual ao PC)
	essentialCmds := []string{
		"APPEND", "BROWSE", "CLEAR / CLS", "CONTINUE", "COPY", "COUNT", "CREATE",
		"DELETE", "DISPLAY", "DO", "EDIT", "FIND", "GOTO / GO", "HELP / ?",
		"INDEX", "LIST", "LOCATE", "MODIFY", "PACK", "QUIT / EXIT", "RECALL",
		"REPLACE", "RESTORE", "SAVE", "SELECT", "SET", "SKIP", "SORT",
		"STORE", "SUM", "TOTAL", "USE", "ZAP", "? / ??", "@ SAY / GET",
	}
	for _, cmd := range essentialCmds {
		if !bytes.Contains(bin, []byte(cmd)) {
			t.Errorf("Catálogo HELP no binário MSX deve listar o comando '%s'", cmd)
		}
	}

	// 3. Verifica presença dos tópicos detalhados de ajuda para comandos
	subtopics := []string{
		"AJUDA: COMANDO USE",
		"AJUDA: COMANDO BROWSE",
		"AJUDA: COMANDO CREATE",
		"AJUDA: COMANDO LIST",
		"AJUDA: COMANDO APPEND",
		"AJUDA: COMANDO EDIT",
		"AJUDA: COMANDO COUNT",
		"AJUDA: COMANDO SUM",
		"AJUDA: COMANDO PACK",
		"AJUDA: COMANDO ZAP",
		"AJUDA: COMANDO CLEAR / CLS",
		"AJUDA: COMANDO QUIT / EXIT",
		"AJUDA: COMANDO INDEX",
		"AJUDA: COMANDO FIND",
		"AJUDA: COMANDO LOCATE",
		"AJUDA: COMANDO REPLACE",
		"AJUDA: COMANDO DELETE",
		"AJUDA: COMANDO RECALL",
		"AJUDA: COMANDO SET",
		"AJUDA: COMANDO ? / ??",
		"AJUDA: COMANDO @ SAY / GET",
		"AJUDA: FUNCOES INTEGRADAS DO DBASE II++",
		"TECLAS DE CONTROLE EM MODO TELA CHEIA (FULL SCREEN)",
	}
	for _, topic := range subtopics {
		if !bytes.Contains(bin, []byte(topic)) {
			t.Errorf("Binário MSX deve conter tópico detalhado de ajuda '%s'", topic)
		}
	}

	// 4. Verifica rotina de escrita direta no console (MSX-DOS BDOS _STROUT 0x09)
	if !bytes.Contains(bin, []byte{0x0E, 0x09, 0xCD, 0x05, 0x00, 0xC9}) {
		t.Errorf("Binário MSX deve conter chamada de escrita direta MSX-DOS BDOS _STROUT (LD C,0x09; CALL 0x0005; RET)")
	}

	// 5. Verifica rotina de inicialização de fonte Graphos III no VDP (SCREEN 0 VRAM 0x0800)
	// Sequência: DI (0xF3), XOR A (0xAF), OUT (0x99), A (0xD3, 0x99), LD A, 0x8E (0x3E, 0x8E), OUT (0x99), A (0xD3, 0x99)
	vdpInitSeq := []byte{0xF3, 0xAF, 0xD3, 0x99, 0x3E, 0x8E, 0xD3, 0x99}
	if !bytes.Contains(bin, vdpInitSeq) {
		t.Errorf("Binário MSX deve conter sequência de configuração do VDP R#14 para inicializar fonte Graphos III")
	}

	// 6. Verifica se os 2048 bytes da fonte Graphos III estão embutidos no binário
	fontData := GetGraphosFontData()
	if len(fontData) != 2048 {
		t.Errorf("GetGraphosFontData deve retornar exatamente 2048 bytes, obteve %d", len(fontData))
	}
	if !bytes.Contains(bin, fontData) {
		t.Errorf("Binário MSX deve conter os 2048 bytes da fonte Graphos III embutidos")
	}
	_ = binStr
}

func TestBuildMSXDOS2BootDisk_AutoexecClean(t *testing.T) {
	tempDir := t.TempDir()
	dskPath := filepath.Join(tempDir, "autoexec_test.dsk")

	err := BuildMSXDOS2BootDisk(dskPath, nil)
	if err != nil {
		t.Fatalf("Erro ao gerar disco para teste de AUTOEXEC.BAT: %v", err)
	}

	dskData, err := os.ReadFile(dskPath)
	if err != nil {
		t.Fatalf("Erro ao ler DSK gerado: %v", err)
	}

	// Localiza AUTOEXEC.BAT no diretório raiz
	rootDirOffset := (ReservedSectors + NumberOfFATs*SectorsPerFAT) * SectorSize
	rootDirBytes := dskData[rootDirOffset : rootDirOffset+RootDirSectors*SectorSize]

	var autoexecCluster uint16
	var autoexecSize uint32
	found := false

	for i := 0; i < RootDirectoryEntries; i++ {
		entry := rootDirBytes[i*32 : (i+1)*32]
		name := string(entry[0:11])
		if name == "AUTOEXECBAT" {
			autoexecCluster = binary.LittleEndian.Uint16(entry[26:28])
			autoexecSize = binary.LittleEndian.Uint32(entry[28:32])
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("AUTOEXEC.BAT não encontrado no diretório raiz do disquete")
	}

	dataAreaOffset := (ReservedSectors + NumberOfFATs*SectorsPerFAT + RootDirSectors) * SectorSize
	fileOffset := dataAreaOffset + int(autoexecCluster-2)*ClusterSize
	autoexecBytes := dskData[fileOffset : fileOffset+int(autoexecSize)]
	autoexecStr := string(autoexecBytes)

	// Não deve ter comandos MSX-BASIC
	for _, forbidden := range []string{"COLOR", "SCREEN", "WIDTH", "CLS"} {
		if bytes.Contains(autoexecBytes, []byte(forbidden)) {
			t.Errorf("AUTOEXEC.BAT não deve conter comando BASIC '%s', conteúdo:\n%s", forbidden, autoexecStr)
		}
	}

	// Deve ter comando MODE 80 e DBASE.COM
	if !bytes.Contains(autoexecBytes, []byte("MODE 80")) {
		t.Errorf("AUTOEXEC.BAT deve conter 'MODE 80', conteúdo:\n%s", autoexecStr)
	}
	if !bytes.Contains(autoexecBytes, []byte("DBASE.COM")) {
		t.Errorf("AUTOEXEC.BAT deve conter chamada 'DBASE.COM', conteúdo:\n%s", autoexecStr)
	}
}

func TestExtractGraphosFont(t *testing.T) {
	// 1. Teste com 2048 bytes puros
	raw := make([]byte, 2048)
	for i := range raw {
		raw[i] = byte(i % 256)
	}
	extractedRaw := extractGraphosFont(raw)
	if !bytes.Equal(extractedRaw, raw) {
		t.Errorf("extractGraphosFont falhou para bloco de 2048 bytes puros")
	}

	// 2. Teste com cabeçalho BSAVE de 7 bytes (formato Graphos III / .ALF: 0xFE, start, end, exec)
	bsave := make([]byte, 7+2048)
	bsave[0] = 0xFE
	binary.LittleEndian.PutUint16(bsave[1:3], 0x0800)
	binary.LittleEndian.PutUint16(bsave[3:5], 0x0FFF)
	binary.LittleEndian.PutUint16(bsave[5:7], 0x0800)
	copy(bsave[7:], raw)

	extractedBSAVE := extractGraphosFont(bsave)
	if !bytes.Equal(extractedBSAVE, raw) {
		t.Errorf("extractGraphosFont falhou ao descartar os 7 bytes de cabeçalho BSAVE")
	}
}
