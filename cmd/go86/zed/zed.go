package zed

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	bios "go86.org/go86/bios"
	cpu "go86.org/go86/cpu"
	dos "go86.org/go86/dos"
)

func DoZed(p string) error {
	f, err := os.Open(p)
	if err != nil {
		return fmt.Errorf("failed to open directory: '%s'; error: %s", p, err)
	}
	defer f.Close()
	fileInfo, err := f.Stat()
	if err != nil {
		return nil
	}

	if fileInfo.IsDir() {
		return zedDir(p)
	}

	// p is a file
	rootDir := filepath.Dir(f.Name())
	name := filepath.Base(f.Name())
	resfile := path.Join(rootDir, "res_"+name)
	return zedFile(f.Name(), resfile)
}

func zedDir(dirName string) error {
	f, err := os.Open(dirName)
	if err != nil {
		return fmt.Errorf("failed to open directory: '%s'; error: %s", dirName, err)
	}
	defer f.Close()
	abspath, err := filepath.Abs(f.Name())
	if err != nil {
		return fmt.Errorf("failed to get absolute path: '%s'; error: %s", dirName, err)
	}
	fmt.Println("Zedning directory:", abspath)
	names, err := f.Readdirnames(-1)
	if err != nil {
		return fmt.Errorf("failed to read directory: '%s'; error: %s", dirName, err)
	}

	for _, name := range names {
		filename := strings.ToLower(path.Base(name))
		if !strings.HasSuffix(filename, ".bin") {
			continue
		}
		if strings.HasPrefix(filename, "res_") {
			continue
		}
		fullpath := path.Join(dirName, name)
		if !slices.Contains(names, "res_"+name) {
			if filename != "jmpmov.bin" {
				return fmt.Errorf("failed to find resource file: 'res_%s'", name)
			} else {
				continue
			}
		}
		resfile := path.Join(dirName, "res_"+name)
		err := zedFile(fullpath, resfile)
		if err != nil {
			fmt.Printf("failed to zed file: '%s'; error: %s\n", name, err)
		}
	}
	return nil
}

func zedFile(filename string, resultsFileName string) error {

	fmt.Println("zed: ", filename, " ", resultsFileName)

	exe, err := dos.ReadExeFromFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file header from: '%s'; error: %s", filename, err)
	}
	// These are binary images not COM files, so no PSP needed nor wanted.
	exe.Etype = dos.IMAGE

	c := cpu.NewCpu(1024 * 1024)
	bios.NewBios(c)
	dos.NewDos(c)
	c.Flags.ReplaceAllFlags(0x02)
	cs := uint(0x1000)
	copy(c.Mem.At(cs, 0), exe.Data)
	c.Regs.SetSeg16(cpu.CS, cs)
	c.Regs.SetSeg16(cpu.DS, 0x0000)

	c.Run()

	return zedValidate(c, resultsFileName)
}

func zedValidate(c *cpu.CPU, resultsFileName string) error {
	resultsFile, err := os.Open(resultsFileName)
	if err != nil {
		return fmt.Errorf("failed to open results file: '%s'; error: %s", resultsFileName, err)
	}
	defer resultsFile.Close()
	resultsFileInfo, err := resultsFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat results file: '%s'; error: %s", resultsFileName, err)
	}
	results := make([]byte, resultsFileInfo.Size())
	_, err = resultsFile.Read(results)
	if err != nil {
		return fmt.Errorf("failed to read results file: '%s'; error: %s", resultsFileName, err)
	}

	errorCount := 0
	var errs []error
	for i, result := range results {
		actual := c.Mem.AbsMem8(i)
		if actual != result {
			errorCount++
			errs = append(errs, fmt.Errorf("memory mismatch at 0x%X: expected 0x%X, got 0x%X", i, result, actual))
		}
	}

	err = errors.Join(errs...)
	if err != nil {
		dumpMem(c.Mem, len(results))
		return err
	}
	fmt.Printf("Zed validation successful for file: %s", resultsFileName)
	return nil
}

func dumpMem(mem *cpu.Memory, len int) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`
Memory dump (len=%d):

    : 00 01 02 03 04 05 06 07 08 09 0A 0B 0C 0D 0E 0F
----: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --`, len))
	start := 0
	for i := start; i < len; i++ {
		if i%16 == 0 {
			fmt.Println(sb.String())
			sb.Reset()
			sb.WriteString(fmt.Sprintf("%04X: ", i))
		}
		sb.WriteString((fmt.Sprintf("%02X ", mem.AbsMem8(i))))
	}
	fmt.Printf("%s\n\n", sb.String())

}
