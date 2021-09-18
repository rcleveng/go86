; Testing add with overflow
org 100h

push cs
pop ds

mov si, banner
call printmsg

mov ax, 0F000h
mov es, ax

cmp word [oper], 0
jz fail

add word [oper], 80h
cmp word [oper], 0
jnz fail
jmp pass

add word [oper2], 80h
add word [oper2], 80h
cmp word [oper2], 0
jz fail
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

banner   db 'Test for ADD with overflow and CMP',13,10, 0

strgood  db 'passed!',13,10,0
strfail  db 'FAILED!',13,10,0     
dot      db '.',0

oper dw 0ff80h
oper2 dw 0h
