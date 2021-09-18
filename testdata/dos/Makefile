AS = nasm

%.com : %.asm
	$(AS) -fbin $< -o $@

BINARIES = testlea.com branch.com smoke.com hello.com addof.com jmps.com

all: $(BINARIES)

clean: 
	-rm $(BINARIES)