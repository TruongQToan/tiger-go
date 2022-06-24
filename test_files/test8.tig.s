	.global main
	.data

	.text
L38:
li t39, 10
li t40, 20
bgt t39, t40, L32
b L33
L33:
li t31, 40
L34:
move $rv, t31
b L37
L32:
li t31, 30
b L34
L37:
