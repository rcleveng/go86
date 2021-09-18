bits 16

;.model small
;.stack 100h


segment code

start:
    ; init data segment, stack segment, and set stack pointer    
    ;mov ax, data
    ;mov ds,ax
    ;mov ax, stack
    ;mov ss,ax
    ;mov sp,stacktop

    mov  al,[ThreeBytes]
    add  al,ThreeBytes+1
    add  al,ThreeBytes+2
    mov  [TheSum] ,al

    mov  ax,4c00h            ; end program
    int  21h

segment data

ThreeBytes db 10h,20h,30h
TheSum     db 00h


segment stack
resb 100h
stacktop: