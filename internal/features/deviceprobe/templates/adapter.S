.syntax unified
.thumb

.section .text.tinygo_scanCurrentStack,"ax",%progbits
.global tinygo_scanCurrentStack
.type tinygo_scanCurrentStack, %function
.thumb_func
tinygo_scanCurrentStack:
    push {r4-r11, lr}
    mov r0, sp
    bl tinygo_scanstack
    add sp, #32
    pop {pc}
.size tinygo_scanCurrentStack, .-tinygo_scanCurrentStack

.section .text.DisableInterrupts,"ax",%progbits
.global DisableInterrupts
.type DisableInterrupts, %function
.thumb_func
DisableInterrupts:
    mrs r0, primask
    cpsid i
    bx lr

.section .text.EnableInterrupts,"ax",%progbits
.global EnableInterrupts
.type EnableInterrupts, %function
.thumb_func
EnableInterrupts:
    msr primask, r0
    bx lr

.section .text.SemihostingCall,"ax",%progbits
.global SemihostingCall
.type SemihostingCall, %function
.thumb_func
SemihostingCall:
    bx lr

.section .text._exit,"ax",%progbits
.global _exit
.type _exit, %function
.thumb_func
_exit:
    b _exit

.section .text._kill,"ax",%progbits
.global _kill
.type _kill, %function
.thumb_func
_kill:
    movs r0, #1
    negs r0, r0
    bx lr

.section .text._getpid,"ax",%progbits
.global _getpid
.type _getpid, %function
.thumb_func
_getpid:
    movs r0, #1
    bx lr
