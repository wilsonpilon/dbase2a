# Makefile gerado para compilação SDCC do projeto MSX TESTE
CC = sdcc
CFLAGS = -mz80 --no-std-crt0 --code-loc 0x0100 --data-loc 0x0000
TARGET = TESTE.COM

all: $(TARGET)

$(TARGET): main.c
	$(CC) $(CFLAGS) main.c -o TESTE.ihx
	hex2bin -e com TESTE.ihx
	cp TESTE.com $(TARGET)
	@echo "Executável MSX-DOS 2 $(TARGET) gerado com sucesso!"

clean:
	rm -f *.ihx *.hex *.bin *.com *.lk *.map *.noi *.rel *.sym *.asm
