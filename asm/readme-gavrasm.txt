Program gavrasm

AVR Assembler for ATMEL AVR-Processors
======================================

Command line Assembler for ATMEL AVR-Processors of the types
- AT90S,
- AT90CAN,
- ATtiny,
- ATmega,
- ATxmega,
- ATA.

For Linux i386-x64 pre-compiled executable : gavrasm
For Win64 compiled executable: gavrasm.exe
Sourcecode for fpc-Pascal

Calls and options
-----------------

Call: gavrasm [-abejlmqswx] SourceFile[.asm]
   Parameters:
     a: Switch output of ANSI-chars on/off
        default is on (Linux) resp. off (Dos, Win)
     b: Extended error comments for beginners
        default is off
     e: Longer error messages
        default is off (short error messages)
     j: Switch PC+X+1 off and add PC+X
        (default is PC+X+1)
     l: Suppress output of the LIST-file
        default is LIST-file output on
     m: Listing of expanded macros on
        default is listing off
     p: Suppress address check in EEPROM segment
        default is check on
     q: No output of messages on the command line (quiet)
        (except: headers and forced messages!)
        default is output on
     s: Output a list of symbols in the LIST-file
        default is no symbol list
     w: Enable wrap-around
        default is wrap-around off
     x: Disable using internal def.inc information
        default is internal information
   Output files:
     SourceFile.lst ... Assembler-Listing, Textfile
     SourceFile.hex ... Code in Intel-Hex-format, Textfile
     SourceFile.eep ... EEPROM-content in Intel-Hex-format, Textfile,
                        (deleted if no .ESEG content was generated)
     SourceFile.err ... Error messages (if any), Textfile

Call: gavrasm [-?hdt]
   Parameters:
     ?: List of options
     h: List of options
     d: List of supported directives
     t: List of supported AVR types

Remark for the option a:
   The Linux version outputs ANSI code sequences by default,
   with other OS's this option is disabled by default. The
   option a inverts this switch under all operating systems.

Result reporting on exiting (%errorlevel%):
   0: Assembled without errors
   1: Assembled with errors
   2: Fatal internal error

Specialties of this Assembler
-----------------------------
Following only the special properties of this assembler are
discussed. All properties not mentioned explicitly, work like
in other assemblers, so these are compatible with others
(e.g. with the ATMEL(R)-assembler).

a) Differences to other assemblers that result in different
   code:
   * Absolute distances written like pc+x or pc-x assume
     that the instruction counter has already been increased
     and the distance +x or -x adds to pc+1. This seems more
     logic to me because the processor works like that and
     because a distance of Zero (pc+0) simply jumps to the
     next instruction. Other assemblers work different and
     this results in different code!
     Just add a "j" to the calling parameters to switch
     this off.

b) Main differences resulting in error and warning messages
   for code that assembles correct (without any errors and
   warnings) with other assemblers:
* Header files for the AVR type (*def.inc) are not necessary,
  gavrasm knows them all by itself. Header files for known
  AVR types, included using the INCLUDE directive, are ignored,
  the file is not read. A warning is given. Instead of reading
  the header file, gavrasm identifies the AVR type from that
  line and, if recognized, defines its own AVR type-specific
  symbol set. Be aware of that in the default mode if you
  change the header file for an AVR, because these changes
  don't come into effect under gavrasm.
  If you want to use a self-written def.inc, then avoid using
  the term "def.inc" in the filename. That prevents gavrasm
  from recognizing the header file and includes it normally.
  This property can be forced by using the parameter x on the
  command line.
* Filenames (e.g. in the INCLUDE directive, usually do not
  need to be in quotes. Under Linux always use quotes
  ("...") for the filename!
* Constant displacements in relative jump- or branch-
  instructions, usually written like " brcc PC+10" (where
  PC stands for PROGRAM COUNTER), are already recognized
  if the displacement is preceeded by a '+' or '-'
  character, like " brcc +10". The use of the notation PC+10
  is optional.
* The extra features in interpreting text and character con-
  stants, implemented in gavrasm, may cause error messages
  and incompatibilites with other assemblers. The differences
  in detail:
  - Introducing a " within a text constant by using a double
    "" is misinterpreted by other assemblers. This causes an
    error message saying "garbage at end of line", the text
    behind the double "" is not interpreted.
  - A ; within a text string ("...;...") or as an ASCII
    character constant (';') is misinterpreted by other
    assemblers as beginning of a comment, they ignore all
    behind the ; and produce an error message.
  - If non-printable ASCII code characters are part of a
    text constant, these are in gavrasm written like
    "...\m\j\@". Other assemblers that do not have this
    feature misinterpret the \ as a normal ASCII character.
    No error message results. If the number of control
    characters is odd, a warning results saying that the
    number of characters is odd and a Null-char is added.
  - In .DB enclosed text is checked for date, time and source
    code keywords. Those are exchanged with the current values
    during assembling. Recognized are:
    %YEAR%, %MONTH%, %DAY%, %HOUR%, %MINUTE%, %SOURCE%
    Please note that those keywords are case sensitive! The
    actual values are now also displayed in the listing.
* The .ERROR directive in gavrasm expects that a text is
  given as parameter, to be displayed as error message. Other
  assembler don't know that. The text makes sense because one
  needs to know what exactly caused the error. To get the same
  effect, code for other assemblers requires the use of an
  additional .MESSAGE directive.
* The instructions BRBC and BRBS previously used single-
  character bit constants like C, Z, N, V, S, H, T and I.
  These bit constants were pre-defined in version 3.2. The
  use of these constants leads to an error message in version
  3.3 and later (undefined symbol).
  Hint: These bit constants are defined in most of the def.inc
  files as SREG_X. Please use those constants instead of the
  single character names. The definitions for AT90S2333,
  ATmega83, ATmega104 and ATmega161C don't include those
  constants.

c) Source code is available and documented:
* FPC-Pascal, written and tested for Linux-(i386), and Win64-FPC
  (use latest version, older versions have problems with some of
  the Pascal instructions)
  (FPC-Compiler available at http://www.freepascal.org).
  Several warnings during compilation can be ignored.
* Source code files provided:
  gavrasm.pas:  Main program file
  gavrdev.pas:  Unit that provides all hardware characteristics
                and symbols of all supported AVR types, *def.inc
                Include files are not needed
  gavrif.pas:   Unit providing nested if/else/endif support
  gavrinst.pas: Processes the instruction set and checks valid
                instructions
  gavrlang.pas: Unit for language support (currently available
                in an english (gavrlang_en.pas), french
                (gavrlang_fr.pas), turkish (gavrlang_tr.pas) and
                german (gavrlang_de.pas) version; to compile your
                language version replace your gavrlang.pas file
                by the desired language version and compile with
                the latest FPC version 
  gavrline.pas: Unit to split the asm lines into pieces
  gavrmacr.pas: Unit for Macro-processing
  gavrout.pas:  Unit for Hex output
  gavrsymb.pas: Unit for symbol-processing
* Prior to compilation: copy the language file gavrlang_en.pas
  and rename it to gavrlang.pas! For the french version use
  gavrlang_fr.pas!
* Test file for checking of the assembler:
  instr.asm:    Test file with all AVR instructions
  Note that this file doesn't specify a certain AVR type, so
  all instructions are valid.

d) Extra directives (Pseudo opcodes):
* DEVICE now automatically provides all symbols of the respective
  AVR type, forget the *def.inc-files, you'll never need them any
  more.
* The EXIT-directive without conditions works normal (interrupts
  further assembling process of the currently processed file)
  and is compatible with other assemblers. If an additional
  parameter is  given, usually a comparison, this is interpreted
  as a boolean value. If this is 1 (TRUE) the EXIT directive is
  executed, providing a conditional stop of the assembling
  process (stops the assembling process completely). This
  conditional EXIT directive also works within INCLUDE files.
  This feature can be used to check for error conditions during
  the assembly process (e.g. stop if limits for certain symbol
  values are exceeded, as an overflow detection).
  This directive can, in part, be exchanged with the directive
  .ERROR, that produces an error message.
* Additional .IF, .ELSE, .ELIF and .ENDIF directive: .IF (condition)
  assembles the following code only if the condition is true (1),
  otherwise branches either to the code after .ELSE, .ELIF or .ENDIF.
  Please use the ELIF directive as an alternative to ELSE, and do
  not mix these, the result is not defined.
  Allows nested .IF, .ELSE, .ELIF and .ENDIF directives of any depth.
* The .IFDEVICE parameter directive compiles the following code,
  if the type specified by the parameter is equal to the one
  specified in the .DEVICE statement.
* .IFDEF and .IFNDEF tests for defined and unsdefined symbols
  and works like the .IF directive.
* .UNDEF undefines previously defined sysmbols.
* .MESSAGE displays a message (in "..." as parameter) on the list
  output, which is not suppressed in the quiet-mode.
* .ERROR forces an error with a message (in "..." as parameter) on
  the list output and in the error textfile.
* Additional .SETGLOBAL p1,p2,[...] directive to export local
  macro symbols. Otherwise symbols defined within a macro are
  completely local and cannot be used outside of the macro. Be
  careful with that feature, it might be source for errors.
* Recursive .INCLUDE-directive with unlimited depth. Under Linux,
  always use quotes for the filename ("...").
* List of supported directives available by gavrasm -d

e) Makros:
* Extended macro capabilities:
  - recursive macro calls without limitations
  - extended parameter checking within the macro definition, not
    only at the macro call itself
  - all labels and symbols in macros are defined locally, they
    are not accessible from outside the macro, improves error
    recognition
  - export of local symbols by use of the SETGLOBAL directive
  - labels in macros are correct even during forward use and if
    used in recursive calls
  - optional use of .ENDM or .ENDMACRO to close the macro definition
  - List of all defined and used macros in the listfile, if the symbol
    list is switched on with -s (and by not suppressing list output
    on the command line by not using the -l option or by .LIST in the
    source code)

f) Improved error detection, commenting and syntax checking:
* Extended symbol checking:
  - exact symbol type checking (cases R for registers, C for
    constants, V for variables, T for AVR type definitions,
    M for local macro labels, L for global macro labels)
  - the use of mnemonics as symbol names is prohibited,
    excludes those used by ATMEL within some older header
    files (accepts OR and BRTS as a symbol name)
  - extended recognition of undefined symbols
* Extended error commenting and warnings (more than 100
  error types, 8 warning types)
* Enhanced syntax commenting for beginners (option b) for
  instructions and directives in the list file
* Extended calculation check (Over- and Underrun of internal
  32-bit-Integer values)
* Extended characters in literal constants and strings: ';',
  '''', '\n', '\\', "abc\n\j", "abc;def"
* Separator character (dots, '.') in binary- and hex-numbers
  for improved readability of the source code (e.g. 0b0001.0010,
  $124A.035F, 0x82.FABC)
* BYTE1-function implemented (similiar to BYTE2 etc.)
* Allows constant displacement in relative jumps, e.g. rjmp +10 or
  brcc -3 (displacement must start with + or - in that case)

g) Supported AVR-types:
* List of supported AVR types with gavrasm -t
* AT90S:    1200, 2313, 2323, 2333, 2343, 4414, 4433, 4434, 8515, 8535
* AT90C85:  34
* AT86RF:   401
* ATtiny:   4, 5, 9, 10, 11, 12, 13, 13A, 15, 20, 22, 24, 24A, 25, 26, 28,
            40, 43U, 44, 44A, 45, 48, 80, 84, 84A, 85, 87, 88, 102, 104,
            167, 212, 214, 261, 261A, 412, 414, 416, 417, 441, 461, 461A,
            814, 816, 817, 828, 840, 841, 861, 861A, 1614, 1616, 1617, 1634,
            2313, 2313A, 3214, 3216, 3217, 4313
* ATmega:   8, 8A, 8HVA, 8U2, 16, 16A, 16HVA, 16HVA2, 16HVB, 16HVBrevB,
            16M1, 16U2, 16U4, 32, 32A, 32C1, 32HVB, 32HVBrevB, 32M1, 32U2,
            32U4, 32U6, 48, 48A, 48P, 48PA, 48PB, 64, 64A, 64C1, 64HVE,
            64HVE2, 64M1, 64RFR2, 83, 88, 88A, 88P, 88PA, 88PB, 103, 104,
            128, 128A, 128RFA1, 128RFR2, 161, 161C, 162, 163, 164A, 164P,
            164PA, 165, 165A, 165P, 165PA, 168, 168A, 168P, 168PA, 168PB,
            169, 169A, 169P, 169PA, 256RFR2, 323, 324A, 324P, 324PA, 324PB,
            325, 325A, 325P, 325PA, 328, 328P, 328PB, 329, 329A, 329P, 329PA,
            406, 640, 644, 644A, 644P, 644PA, 644RFR2, 645, 645A, 645P,
            649, 649A, 649P, 1280, 1281, 1284, 1284P, 1284RFR2, 2560, 2561,
            2564RFR2, 3208, 3209, 3250, 3250A, 3250P, 3250PA, 3290, 3290A,
            3290P, 3290PA, 4808, 4809, 6450, 6450A, 6450P, 6490, 6490A,
            6490P, 8515, 8535
* ATxmega:  8E5, 16A4, 16A4U, 16C4, 16D4, 16E5, 32A4, 32A4U, 32C3, 32C4,
            32D3, 32D4, 32E5, 64A1, 64A1U, 64A3, 64A3U, 64A4U, 64B1, 64B3,
            64C3, 64D3, 64D4, 128A1, 128A1U, 128A3, 128A3U, 128A4U, 128B1,
            128B3, 128C3, 128D3, 128D4, 192A3, 192A3U, 192C3, 192D3, 256A3,
            256A3B, 256A3BU, 256A3U, 256C3, 256D3, 384C3, 384D3
* AT90CAN:  32, 64, 128
* AT90PWM:  1, 2, 2B, 3, 3B, 81, 161, 216, 316
* AT90USB:  82, 162, 646, 647, 1286, 1287
* ATA:      5272, 5505, 5700M322, 5702M322, 5781, 5782, 5783, 5787, 5790,
            5790N, 5791, 5795, 5831, 5832, 5833, 5835, 6285, 6286, 6289,
            6612C, 6613C, 6614Q, 6616C, 6617C, 8210, 8215, 8510, 8515, 664251

h) Language versions currently supported:
* English, German, French, Turkish (just copy the file avrlang_xx.pas to
  avrlang.pas by overwriting it and recompile with fpc)

i) Open issues and unresolved errors:
* For undefined reasons ATMEL uses several C-type definitions in its def.inc
  files. These are insufficiently implemented when reading those files to
  generate symbol lists. Unless I receive an error message from a user I will
  not implement those in gavrasm completely.

Versions and Changes
--------------------
Latest versions at http://www.avr-asm-tutorial.net/gavrasm/index_en.html

Hint: gavrasm's latest version is included in the AVR assembler
  simulator which is available as Lazarus source code as well as
  in compiled executable versions for Linux and Windows 64 bit here:
  http://www.avr-asm-tutorial.net/avr_sim/index_en.html

December 2018: Version 4.1
- Corrected: Constants with the names C, Z, N, H, I, T, V and S were not
  displayed in the symbol list. This is corrected.
- Corrected: When calling the functions HIGH, LOW, BYTE1 to BYTE4, LWORD
  and HWORD with negative numbers those were not correctly executed.
- Corrected: When checking negative numbers those were not correctly 
  testet for their lower limit -128, and the necessary error message
  was not issued.
- Corrected: In nested .IF's, when .ENDIF of the inner condition was
  reached within the .ELSE part of the outer condition, an error
  occurred, because the initial status of the outer condition was
  inverted.

September 2018: Version 4.0
- Corrected: A problem with leading zeroes in constants (0123) has been
  fixed.
- Corrected: A missing closing bracket in a function lead to a
  missleading error message.
- Corrected: The reference to a forward label lead in pass 1 to an
  error message. 

July 2018: Version 3.9
- Added: Operator % (modulo, residue) added.
- Corrected: With the instructions CPI, LDI gavrasm allowed to not only
  use a constant as second parameter but allowed a register, too. By
  falsely using the register's address to compare or load. This
  has been changed.
- Corrected: Expressions such as a-b*c yielded false results.
- Changed: The device databooks for several Tiny-, Mega- and
  XMega-Devices is available, for those types instruction
  checks are switched on.  

April 2018: Version 3.8
- Corrected: Range Check Error when ending path with macro definition
  still open, checked at end of pass 1 and error message added
- Corrected: Crash in gavrout when hex line checksum is zero
- Corrected: LD/ST- and ADIW/SBIW-instructions in ATtiny441/841
- Extended: Range Check Error when number of parameters per line
  exceeds 50 (e.g. in a table). Extended to 100 parameters
- Added: Support for ATmega3208/09/4808/09 (no datasheet available
  yet)

October 2017: Version 3.7
- Corrected: EEPROM segment that started with .ORG larger than 0 caused
  a wrong address in the hex file, error has been corrected
- Corrected: Fatal error when using more than one .cseg directives
- Corrected: A minor error in the symbol routine
- Corrected: Closing bracket without opening bracket in parameter led to
  range check error
- Corrected: Error when using relative branches PC+/- or +/- corrected,
  option J removed
- Changed: The parameter -m was described in wrong direction in previous
  ReadMe files: the default is macro expansion listing off
- Changed: The parameters %YEAR% etc. are now displayed as they actually
  are in the listing
- Added: Support for several newer AVR types

Januar 2017: Version 3.6
- Added: Support for ATtiny80 und 840 (datasheet not yet available),
  ATtiny102, 104, 417, 814, 816 und 817, def.inc as of Studio version
  7.0-1188
- Corrected: Error in device memory sizes of majure types

December 2015: Version 3.5
- Added: Support for devices (Studio Version 7.0.634)
  ATA: 5702M322, 5781, 5782, 5783, 5791, 8210, 8215, 8510, 8515
  ATxmega: 32C3, 32D3
  ATmega: 168PB, 16hvbrevb, 324PB, 328PB, 48PB, 88PB
  ATtiny: 441, 841
- Added: Use of date-, time- and source-keywords in .DB text, recognized are:
    %YEAR%, %MONTH%, %DAY%, %HOUR%, %MINUTE%, %SOURCE%
  Please note that those keywords are case sensitive!
- Added: -p option to suppress EEPROM address error checking

December 2014: Version 3.4
- Added: Return codes at program end (0 no errors, 1 one or more, 2 fatal error)
- Added: Syntax check for ATtiny828
- Added: 28 new device types
- Changed: Unhandled exception if an unknown device is selected

June 2012: Version 3.3
- Added: Documentation for a few types added
- Changed: Standard symbols encoded more efficient, reduced exe by 32%
  and unit gavrdev by 82%
- Changed: AVR type output with instruction set flags

May 2012: Version 3.2
- Added: Turkish language version
- Added: Support for ATxmega128A1U, 128A4U, 128C3, 128D4, 16C4, 192C3, 256C3,
  32C4, 384D3, 64A1U, 64A4U, 64C3, 64D4, AT90PWM161, ATtiny828, AT90S2333,
  ATmega161C (Compatibility mode)
- Added: ELPM instructions valid for ATmega1284xx, even though those are
  not listed in the current datasheet, but are obviously implemented in the
  chips and work fine
-Added: ADIW and SBIW instructions now allow register pair syntax such as
  ADIW R31:R30,C

December 2011: Version 3.1
- Added: Support for five new XMEGA types (status: Studio 5.0.516)
- Corrected: Error in instruction set checking
- Added: Ignores C style lines in include files

July 2011: Version 3.0
- Corrected: Parser now allows empty lines with blanks and
             tabs only, without using ';' as "would be correct"
- Added: Support for 10 new AVR types (status: Studio 5)
- Added: Instructions DES, XCH, LAS, LAC, LAT

January 2011: Version 2.9
- Corrected: Error in processing constants
- Corrected: Error in addressing SRAM in AVR8L types 

January 2011: Version 2.8
- Added: Support for six new AVR types (status: Studio 4 SP3)
- Changed: New Documentation for ten types (Status: 10/2010)
- Corrected: Errors in register checking and LDS/STS for the AVR8L types
             ATtiny4, 5, 9, 10, 20 and 40

November 2010: Version 2.7
- Corrected: Error in symbol list

May 2010: Version 2.6
- Corrected: Type conversions for variables in Pascal source code
- Corrected: Definition of double registers with external def.inc
- Corrected: Errors when using literal constants in .db and with \
- Corrected: Detection of erroneous code lines 
- Added: 6 new AVR types

February 2010: Version 2.5
- Corrected: Serious error in checking for type-specific instructions 

January 2010: Version 2.4
- Corrected: Error in symbol checking
- Added: 24 additional AVR types
- Added: .UNDEF directive
- Changed: Coding of the instruction sets

May 2009: Version 2.3
- Corrected: Serious error when addressing register ZL (R30).
- Corrected: Error when using the -x option (only external def.inc) 

December 2008: Version 2.2
- Added: Support for french version (thanks to Herve)
- Added: 65 additional AVR types
- Changed: The whole internal storage and processing of hardware-
  specific properties of the AVRs, the def.inc files and the type-
  specific internal symbols was re-worked and changed.
- Removed: The support for AVR types AT90C8534, AT90S2343, ATmega104
  and ATmega603 was removed (these types are no more supported by
  ATMEL - obsolete devices).
- Added: SPM Z+ support.

November 2006: Version 2.1
- Corrected: Bug in CALL and JUMP opcodes
- Corrected: Recognition of illegal characters in instructions

September 2006: Version 2.0
- Corrected: Bug in handling variables with .SET

August 2006: Version 1.9
- Corrected: Recognition of already defined symbols corrected
- Corrected: Incorrect opcode of STD without displacement

July 2006: Version 1.8
- Corrected: False memory adresses of internal SRAM in some ATmega

May 2006: Version 1.7
- Corrected: ADIW/SBIW check changed
- Corrected: Re-worked the ELIF directive

December 2005: Version 1.6
- Added: Support for types AT90CAN32 and 64, ATtiny24, 44 and 84,
  ATmega 644

September 2005: Version 1.5
- Corrected: Double list output in directives removed.
- Corrected: Problem with nested IF/IFDEF/IFNDEF/IFDEVICE
  directives and undefined variables resolved.

July 2005: Version 1.4
- Added: Support for types AT90CAN128, ATmega329, 3290, 406,
  649, 6490, AT90PWM2, 3
- Changed: Unlike in earlier versions, directives are allowed in
  columns >1 and instruction may start in column 1 of a line.
- Changed: Within the D- and E-segment, now all directives are
  allowed.
- Corrected: An error while working with the constant ','.
- Corrected: Device instruction did not recognize some AVR types
  with longer names.
- Corrected: Port ADCSRA in ATmega8 was only recognized under its
  former name.

April 2005: Version 1.3
- Corrected: EEPROM capacity of the ATmega104 and 128
- Added: Support for types ATmega 1280, 1281, 2560, 2561 and 640

March 2005: Version 1.2
- Corrected: Error in bit WGM02 of port TCCR0B in the ATtiny13
  constants
- Corrected: Missing parameters in a number of instructions were
  accepted without an error
- Added: Support for the new types ATtiny25, 45 and 85

January 2005: Version 1.1
- Corrected: Error in the number of warnings
- Corrected: Error in using SETGLOBAL
- Corrected: Error in using underscore in macro names
- Added: Support for using the currect program counter PC as a variable,
  e.g. in RJMP PC+3 or brcc PC-2.

October 2004: Version 1.0
- Added: Support for new types ATmega325/3259/645/6450
- Changed: Complete rewriting of the directives
- Added: New directives .ERROR, .MESSAGE, .ELIF, .IFDEF and .IFNDEF
  for enhanced compatibility with the ATMEL assembler
- Added: Error output to a separate error file source.err.

March 2004: Version 0.9
- Added: Support for the new types ATmega48/88/168

Febrary 2004: Version 0.8
- Corrected: Error with setting the symbol RAMEND removed
- Corrected: Some SRam settings for older ATmega types were false
- Corrected: Assembler did not run in the short error message mode
  by default

October 2003: Version 0.7
- Added: Support for new type ATtiny2313
- Added: .IFDEVICE directive
- Changed: Reliability of ATtiny22 instruction set now fine

September 2003: Version 0.6
- Corrected: Error when using negative numbers in binary operations
- Corrected: Error with type checking in .DEVICE statement
- Added: def.inc include files now trigger the internal symbol defini-
  tions, no error message any more, a warning is given and the def.inc-
  file is not read
- Changed: Complete instruction set reworked
- Added: Several older device types, warning for undocumented types
- Added: Nested .IF, .ELSE and .ENDIF directives

August 2003: Version 0.5
- Corrected: LD/ST instruction for type AT90S1200 added,
  resulted in an error message in previous versions

July 2003: Version 0.4
- Corrected: Misleading error message when using an undefined
  register
- Added: DEVICE now adds all predefined symbols, as listed in
  the data sheet of the respective processor (these were
  previously provided by def.inc-files). Attempting to include
  a *def.inc-file results in an error message!
- Added: Support for the new type AVRtiny13 (even before ATMEL
  has released its def.inc!)

June 2003: Version 0.3
- Corrected: Detection of double labels improved (ran into
  internal compiler error)
- Corrected: Constant expressions starting and ending with a
  bracket lead to a strange error message, therefore removed
  the support for macro calls with parameters in brackets

May 2003: Version 0.2
- Added: SETGLOBAL directive for export of local symbols in
  macros
- Added: Wrap-around allowed in relative jumps/calls. Wrap is
  also possible in relative (conditional) branches (BRCC etc.),
  difference to the ATMEL assembler!
- Fixed: Addressing in hex-file now compatible with Studio/Wavrasm
  (byte-oriented, not word-oriented, also handles multiple .ORG
  directives correct
- Fixed: LSL and LSR instruction for AT90S1200 are valid, invalid
  error message in version 0.1
- Fixed: Labels, variables and symbols starting with an underscore
  are now accepted
- Fixed: Double division x/y/z yielded false result, corrected
- Fixed: Instructions starting in column 1 of a line produced a
  misleading error message, added hint
- Fixed: directives that don't start in column 1 of the line but are
  preceded by a label were not recognized, error message added
- Added: Command line option E for shorter (default) or longer
  error messages

December 2002: Version 0.1 Beta (first release)

Terms of use of this software
-----------------------------

- Copyright for all versions: (C)2002..2018 by Gerhard Schmidt
- Free use of the source code and the compiled versions for non-
  commercial purposes. Distribution allowed if the copyright
  information is included.
- No warranty for correct functioning of the software.
- Report errors (especially any compiler errors) and your
  most-wanted features to gavrasm(at)avr-asm-tutorial.net (Subject
  "gavrasm 4.0"). If errors occur, please include the source code,
  and any included files, the listing and the error file.
