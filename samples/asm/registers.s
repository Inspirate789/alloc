#include "textflag.h"

TEXT ·getSP(SB),NOSPLIT,$0-8
MOVQ SP, ret+0(FP)
RET
