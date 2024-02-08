#include <stdio.h>
#include <linux/lirc.h>

void main() {
  printf("LIRC_GET_FEATURES = %#010x\n",LIRC_GET_FEATURES);
  printf("LIRC_CAN_REC_MODE2 = %#010x\n",LIRC_CAN_REC_MODE2);
  printf("LIRC_CAN_SEND_PULSE = %#010x\n",LIRC_CAN_SEND_PULSE);
  printf("LIRC_CAN_SET_SEND_CARRIER = %#010x\n",LIRC_CAN_SET_SEND_CARRIER);
  printf("\n");
  printf("LIRC_SET_REC_TIMEOUT_REPORTS = %#010x\n",LIRC_SET_REC_TIMEOUT_REPORTS);
  printf("LIRC_SET_SEND_CARRIER = %#010x\n",LIRC_SET_SEND_CARRIER);
  printf("LIRC_GET_SEND_MODE = %#010x\n",LIRC_GET_SEND_MODE);
  printf("LIRC_SET_SEND_MODE = %#010x\n",LIRC_SET_SEND_MODE);
  printf("LIRC_MODE_PULSE = %#010x\n",LIRC_MODE_PULSE);
}
