

with open("roms/out.ch8", "wb") as f:
    instructions = [

            #0x61FF, # Store value in register
            #0x62EE, # Store value in register
            #0x63DD, # Store value in register
            #0x6f22, # Store value in register
            #0xA100, # Store 0x100 in I 
            #0xF655,
        
        0x610E, # Store value in register 0
        0x620E, # Store value in register 1
        0xA000, # Store 0x000 in I 
        0xD12F, # Draw sprite consisting of 5 rows at R0, R1
    ]

    for instruction in instructions:
        # Each instruction is a 16-bit value, so we need to convert it to bytes
        f.write(instruction.to_bytes(2, byteorder='big'))
    
