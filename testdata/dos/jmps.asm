; Testing add with overflow
org 100h

push cs
pop ds

mov si, banner
call printmsg

mov ah, 0ffh
cmp ah, ah

ja fail

mov si, srcbigger
mov ax, 0x1000
mov bx, 0x2001

cmp ax, bx
ja fail

sub ax, bx

jmp pass

finished:
ret

pass:
mov si, strgood
call printmsg
ret

fail:
mov si, strfail
call printmsg
jmp finished

printmsg:
mov ah, 0Eh
cld
lodsb
cmp al, 0
jz done
int 10h
jmp printmsg
done:
ret

banner   db 'Test for JMPS',13,10,0
srcbigger db 'Test for Source Bigger',13,10,0

strgood  db 'passed!',13,10,0
strfail  db 'FAILED!',13,10,0     
dot      db '.',0

oper dw 0ff80h
oper2 dw 0h
