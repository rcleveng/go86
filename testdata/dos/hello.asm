; hello2-DOS.asm - single-segment, 16-bit "hello world" program
;
; This demonstrates single-character output as well as string output
; via DOS services
;
; assemble with "nasm -f bin -o hi.com hello2-DOS.asm"

    org  0x100        ; .com files always start 256 bytes into the segment

    ; int 21h is going to want...

    mov  dx, msg      ; the address of or message in dx
    mov  ah, 9        ; ah=9 - "print string" sub-function
    int  0x21         ; call dos services

    mov  dl, 0x0d     ; put CR into dl
    mov  ah, 2        ; ah=2 - "print character" sub-function
    int  0x21         ; call dos services

    mov  dl, 0x0a     ; put LF into dl
    mov  ah, 2        ; ah=2 - "print character" sub-function
    int  0x21         ; call dos services

    mov  ah, 0x4c     ; "terminate program" sub-function
    int  0x21         ; call dos services

    msg  db 'Hello again, World!$'   ; $-terminated message