# rsp2dwarf

This tool acts a replacement to rsp2elf but writes the debugging symbols using dwarf. This allows the resulting .o file to be compatible with gdb.

## Example usage

```bash
# assemble the rsp routine
rspasm -o bin/rsp/microcode rsp/microcode.s
# this line will generate the following files
# bin/rsp/microcode     (IMEM)
# bin/rsp/microcode.dat (DMEM)
# bin/rsp/microcode.dbg (Symbol table)
# bin/rsp/microcode.sym (Source code mapping)

# Generate elf file without debugging symbols
# this also generate linking symbols for rspRoutineTextStart,
# rspRoutineTextEnd, rspRoutineDataStart, rspRoutineDataEnd
# as indicated by the -n flag 
rsp2dwarf bin/rsp/microcode -o bin/rsp/microcode.o -n rspRoutine

# You can also tell the tool to output debugging symbols using the -g flag
rsp2dwarf bin/rsp/microcode -o bin/rsp/microcode.debug.o -n rspRoutine -g

```

Using the above commands you end up with a bin/rsp/microcode.o and a bin/rsp/microcode.debug.o file. Link the bin/rsp/microcode.o file with your ROM. The RSP routines can be accessed from C code.

```C

extern char rspRoutineTextStart[];
extern char rspRoutineTextEnd[];

extern char rspRoutineDataStart[];
extern char rspRoutineDataEnd[];

void startRSPTask() {
    OSTask task;
    ...
    task.t.ucode = (u64*)rspRoutineTextStart;
    task.t.ucode_size = rspRoutineTextEnd - rspRoutineTextStart;
    ...
    task.t.ucode_data = (u64*)rspRoutineDataStart;
    task.t.ucode_data_size = rspRoutineDataEnd - rspRoutineDataStart;
    ...

    osSpTaskStart(&task);
}

```

## Including debugging symbols

It is a good idea to generate a separate file with debug symbols by including the `-g` flag since you don't want the debug symbols to be located where the RSP program is stored in the ROM image. Instead the debugging symbols should be located at 0-0x1000 for the IMEM and 0x04000000 for DMEM. You can then include the debug symbols separately at those fixed addresses using the following GDB command.

```
add-symbol-file bin/rsp/microcode.debug.o -s .text 0x00000000 -s .data 0x04000000
```
