# tarmanip

Tarmanip is a tool to reproducibly and declaratively manipulate tar failes.

It was written to sit within a sqfs2tar -> tar2sqfs pipeline for modifying firmware files, but can be repurposed for other uses.

## Usage

All tarmanip behaviour is defined in a script. The script is a [prototext](https://github.com/google/nvidia_libs_test/blob/master/cudnn_benchmarks.textproto) file, containing a series of changes to apply to the tarball.

For example:

    cat > passwd-modified.txt <<EOF
    root:x:0:0:System administrator:/root:/bin/bash
    EOF
    cat > script.pb.text << EOF
    change { write {
      path: "/etc/passwd"
      source: "./passwd-modified.txt"
    } }
    EOF

For more information about supported changes, see [proto/manipulate.proto](proto/manipulate.proto). The script is very low-level, and it it recommended you use your favourite programming language to generate it from some higher-level description of intent.

Once a script exist, tarmanip can be called to modify any tarball given this script in a reproducible fashion:

    tarmanip -in foo.tar -script script.pb.text -out foo-modified.tar

Tarmanip will strive to introduce the minimum amount of changes in the input tarball, and to not introduce any unhermetic information from the ambient execution environment (eg. timestamps, etc). 

## Usage in vraytekdigor

lib.nix contains makeCustomFirmware which applies a tarmanip against a vendor firmware to build a custom firmware.

The default firmware defined in default.nix generates a simple script at a fairly low level, and uses it to craft a firmware using the above makeCustomFirmware function.
