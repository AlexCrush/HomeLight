;.device ATTiny4313
;.include "tn2313def.inc" ; for Avr Assembler 2
.include "tn4313def.inc" ; for Avr Assembler 2
;-----------------------------------------------------------------------
; TODO
; 1. флаги запрета для каждого выключателя
; 2. двойное нажатие
; 3. игнор включенных выключателей на старте
; 4. оптимизация протокола
; 5. поддержка герконов?
;------------------------------------------------------------------------
.equ ConfigId_Master = 4

; 0 = fixed master (ConfigId_Master) + slaves
; 1 = ring token protocol (README: "План изменения протокола")
.equ UseRingProtocol = 0

.equ ConfigId = 1 ; Котельная
;.equ ConfigId = 2 ; Кладовка
;.equ ConfigId = 3 ; СУ2
;.equ ConfigId = 4 ; Гостиная
;.equ ConfigId = 5 ; СУ2-упр
;.equ ConfigId = 6 ; Детская

;.equ ConfigId = 100 ; debug

.if ConfigId == 1;   Котельная
.equ UseRadio = 1
.else
.equ UseRadio = 0
.endif

.include "hw.asm-inc"
.include "macroses.asm-inc"
.include "consts.asm-inc"
;------------------------------------------------------------------------
.if UseRadio == 1
.equ _Timer1Cmp = Radio_Tick
.else
.equ _Timer1Cmp = _RESET
.endif

.CSEG
.ORG 0x0000
    rjmp _RESET         ;  0
    rjmp _RESET         ;  1 INT0
    rjmp _RESET         ;  2 INT1
    rjmp _RESET         ;  3 TIMER1 CAPT   = Timer/Counter1 Capture Event
    rjmp _Timer1Cmp     ;  4 TIMER1 COMPA  = Timer/Counter1 Compare Match A
    rjmp _RESET         ;  5 TIMER1 OVF    = Timer/Counter1 Overflow
    rjmp _RESET         ;  6 TIMER0 OVF    = Timer/Counter0 Overflow
    rjmp _RESET         ;  7 USART0, RX    = USART0, Rx Complete
    rjmp _RESET         ;  8 USART0, UDRE  = USART0 Data Register Empty
    rjmp _RESET         ;  9 USART0, TX    = USART0, Tx Complete
    rjmp _RESET         ;  A ANALOG COMP   = Analog Comparator
    rjmp _RESET         ;  B PCINT0        = Pin Change Interrupt Request 0
    rjmp _RESET         ;  C TIMER1 COMPB  = Timer/Counter1 Compare Match B
    rjmp _Timer0Cmp     ;  D TIMER0 COMPA  = Timer/Counter0 Compare Match A
    rjmp _RESET         ;  E TIMER0 COMPB  = Timer/Counter0 Compare Match B
    rjmp _RESET         ;  F USI START     = USI Start Condition
    rjmp _RESET         ; 10 USI OVERFLOW  = USI Overflow
    rjmp _RESET         ; 11 EE READY      = EEPROM Ready
    rjmp _RESET         ; 12 WDT OVERFLOW  = Watchdog Timer Overflow
    rjmp _RESET         ; 13 PCINT1        = Pin Change Interrupt Request 1
    rjmp _RESET         ; 14 PCINT2        = Pin Change Interrupt Request 2

_RESET:
    cli
    ldi tmp, LOW(RAMEND)
    out SPL, tmp
    ldi tmp, HIGH(RAMEND)
    out SPH, tmp

    cbi ACSR, ACD ; disable Analog Comparator

    rcall Lamp_Init
    rcall Relay_Init
    rcall Inputs_Init

    sbi DD_RS485DIR, BIT_RS485DIR       ; Out
    cbi Port_RS485DIR, BIT_RS485DIR     ; RS485 - In

    sbi DD_RS485TX, BIT_RS485TX ; Out
    cbi DD_RS485RX, BIT_RS485RX ; In
    sbi Port_RS485RX, BIT_RS485RX ; Pullup

    sbi DD_Led, BIT_Led
    cbi PORT_Led, BIT_Led


; USART config: 57600, 8 data bits, 2 stop bits, NO parity
; stty -F /dev/ttyUSB0 57600 cs8 cstopb -parenb
; nc -l 1234 </dev/ttyUSB0 >/dev/ttyUSB0
; nc localhost 1234
.if 1
    ldi tmp, HIGH(RS485_UBRR)
    out UBRRH, tmp
    ldi tmp, LOW(RS485_UBRR)
    out UBRRL, tmp
    ldi tmp, (1<<USBS)|(3<<UCSZ0) ; Set frame format: 8data, 2stop bit
    out UCSRC,r16
    ldi  tmp, (1<<RXEN)|(1<<TXEN) ; Enable receiver and transmitter
    out UCSRB,tmp
.endif

; Timer0 Config
    ldi tmp, (FREQ_MHZ * 1000000 / 1024) / 50
    out OCR0A, tmp
    ldi tmp, 0b00000010   ; CTC mode
    out TCCR0A, tmp
    ldi tmp, 0b00000101  ; // ClockSource =  Fclk / 1024
    out TCCR0B, tmp      ; enable interrupt on compare match
    in tmp, TIMSK
    ori tmp, 1 << OCIE0A
    out TIMSK, tmp

    clr tmp
    sts Uptime, tmp
    sts Communicated, tmp

.if UseRingProtocol == 1
    rcall LinkRing_Init
.endif

.if UseRadio == 1
    rcall Radio_Init
.endif

.if 0
l1:
    sbi Port_RS485DIR, BIT_RS485DIR
    sbi Port_RS485TX, BIT_RS485TX
    sbi Port_Led, BIT_Led
    rcall WAIT_1_Sec
    rcall WAIT_1_Sec
    rcall WAIT_1_Sec
    cbi Port_RS485TX, BIT_RS485TX
    cbi Port_Led, BIT_Led
    rcall WAIT_1_Sec
    rcall WAIT_1_Sec
    rcall WAIT_1_Sec
    rjmp l1
.endif

    sei
    rjmp LinkLoop
;----------------------------

_Timer0Cmp:
    push tmp
    in tmp, SREG
    push tmp
    push zL
    push zH
    push xL
    push xH
    push yL
    push yH
    push tmp1
    push tmp2
    push tmp3

    lds tmp, Uptime
    inc tmp
    sts Uptime, tmp
    andi tmp, 0b0000111
    brne SkipCheckComminicated
    lds tmp1, Communicated
    tst tmp1
    breq NotCommunicated
    sts Communicated, tmp

    sbi PIN_Led, BIT_Led

NotCommunicated:
SkipCheckComminicated:

    rcall Inputs_Poll
    rcall Relay_CalcAndWrite

    lds tmp, LampOffSubtimer
    dec tmp
    brne SkipLampOffTimer
    rcall Lamp_TickOffTimers
    ldi tmp, LampOffSubtimerMax
SkipLampOffTimer:
    sts LampOffSubtimer, tmp

.if UseRadio == 1
    rcall Radio_ProcessInputs
.endif

.if UseRingProtocol == 1
    rcall LinkRing_OnTick
.endif

    pop tmp3
    pop tmp2
    pop tmp1
    pop yH
    pop yL
    pop xH
    pop xL
    pop zH
    pop zL
    pop tmp
    out SREG, tmp
    pop tmp

    reti

;--------------------------------------------------------------------------------------

;---------------------------------------------------------------------
.include "delays.asm-inc"
.include "relay.asm-inc"
.include "lamp.asm-inc"
.include "inputs.asm-inc"
.include "config-works.asm-inc"
.include "link_l2.asm-inc"

.if UseRingProtocol == 1
.include "link_l7_ring.asm-inc"
.else
.if ConfigId == ConfigId_Master
.include "link_l7_master.asm-inc"
.else
.include "link_l7_slave.asm-inc"
.endif
.endif

.if UseRadio == 1
.include "radio.asm-inc"
.endif
.include "config-vars.asm-inc"
.include "vars.asm-inc"
;---------------------------------------------------------------------
