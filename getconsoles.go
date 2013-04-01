// 1 september 2011
package main

type Consoles map[string]string

func GetConsole(which string) Consoles {
	if what, ok := config.Consoles[which]; ok {
		return Consoles{
			which:	what,
		}
	}
	return nil
}

//var omitConsoles = map[string]bool{
	// TODO
//	"Aurora":						true,		// not sure what it uses (need to ask Nik)
//	"G80":						true,		// used CPU boards
//	"Europa-R":					true,		// not sure what it uses (too new)
//	"Hikaru":						true,		// not sure what type of ROM board it uses
//	"System E":					true,		// not sure what removable media it used but it definitely used removable media
//}
