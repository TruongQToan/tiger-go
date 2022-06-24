	.global main
	.data
L31: .asciiz "Hello, World!
"

	.text
L35:
move t36, $t0
move t37, $t1
move t38, $t2
move t39, $t3
move t40, $t4
move t41, $t5
move t42, $t6
move t43, $t7
move t44, $t8
move t45, $t9
la t46, print
lw t47, 0($fp)
move $a0, t47
la t48, L31
move $a1, t48
jalr t46
move $rv, $rv
b L34
L34:
