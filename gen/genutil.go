package gen

import(
	"bufio"
	"bytes"
	"os"
   "path/filepath"
	"strings"
   "unicode"

	"github.com/ardielle/ardielle-go/rdl"
)

func OutputWriter(outdir string, name string, ext string) (*bufio.Writer, *os.File, string, error) {
   sname := "anonymous"
   if strings.HasSuffix(outdir, ext) {
      name = filepath.Base(outdir)
      sname = name[:len(name)-len(ext)]
      outdir = filepath.Dir(outdir)
   }
   if name != "" {
      sname = name
   }
   if outdir == "" {
      return bufio.NewWriter(os.Stdout), nil, sname, nil
   }
   outfile := sname
   if !strings.HasSuffix(outfile, ext) {
      outfile += ext
   }
   path := filepath.Join(outdir, outfile)
   f, err := os.Create(path)
   if err != nil {
      return nil, nil, "", err
   }
   writer := bufio.NewWriter(f)
   return writer, f, sname, nil
}

func FormatBlock(s string, leftCol int, rightCol int, prefix string) string {
   if s == "" {
      return ""
   }
   tab := Spaces(leftCol)
   var buf bytes.Buffer
   max := 80
   col := leftCol
   lines := 1
   tokens := strings.Split(s, " ")
   for _, tok := range tokens {
      toklen := len(tok)
      if col+toklen >= max {
         buf.WriteString("\n")
         lines++
         buf.WriteString(tab)
         buf.WriteString(prefix)
         buf.WriteString(tok)
         col = leftCol + 3 + toklen
      } else {
         if col == leftCol {
            col += len(prefix)
            buf.WriteString(prefix)
         } else {
            buf.WriteString(" ")
         }
         buf.WriteString(tok)
         col += toklen + 1
      }
   }
   buf.WriteString("\n")
   emptyPrefix := strings.Trim(prefix, " ")
   pad := tab + emptyPrefix + "\n"
   return pad + buf.String() + pad
}

func Spaces(count int) string {
   return StringOfChar(count, ' ')
}

func StringOfChar(count int, b byte) string {
   buf := make([]byte, 0, count)
   for i := 0; i < count; i++ {
      buf = append(buf, b)
   }
   return string(buf)
}

func Capitalize(text string) string {
   return strings.ToUpper(text[0:1]) + text[1:]
}

func SnakeToCamel(name string) string {
   // "THIS_IS_IT" -> "ThisIsIt"                                                                                                    
	result := make([]rune, 0)
	newWord := true
	for _, c := range name {
		if c == '_' {
			newWord = true
		} else if newWord {
			result = append(result, unicode.ToUpper(c))
			newWord = false
		} else {
			result = append(result, unicode.ToLower(c))
		}
	}
	s := string(result)
	return strings.Replace(strings.Replace(s, "Uuid", "UUID", -1), "Uri", "URI", -1)
}

func addFields(reg rdl.TypeRegistry, dst []*rdl.StructFieldDef, t *rdl.Type) []*rdl.StructFieldDef {
   switch t.Variant {
   case rdl.TypeVariantStructTypeDef:
      st := t.StructTypeDef
      if st.Type != "Struct" {
         dst = addFields(reg, dst, reg.FindType(st.Type))
      }
      for _, f := range st.Fields {
         dst = append(dst, f)
      }
   }
   return dst
}

func FlattenedFields(reg rdl.TypeRegistry, t *rdl.Type) []*rdl.StructFieldDef {
   return addFields(reg, make([]*rdl.StructFieldDef, 0), t)
}
