The wasm build is just following: https://ebiten.org/documents/webassembly.html

The gopherjs build is following: https://ebiten.org/documents/gopherjs.html with
help from the main instructions at https://github.com/gopherjs/gopherjs. Because
as of now it still uses go 1.12, I also used xerrors package to have the .Is
error handling.
