package core

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/Velocidex/go-magic/magic"
	"github.com/Velocidex/go-magic/magic_files"
	"github.com/gabriel-vasile/mimetype"
)

func Magic(fpath string) string {
	handle := magic.NewMagicHandle(magic.MAGIC_NONE)
	defer handle.Close()
	magic_files.LoadDefaultMagic(handle)
	return handle.File(fpath)
}

type FileHashFunc func(fpath string) string

func MD5HashFunc(fpath string) (string, error) {
	f, err := os.Open(fpath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

type LsRecord struct {
	MD5       string
	Timestamp int64
	Size      int64
	Path      string
	MimeType  string
	MagicStr  string
}

func (lr *LsRecord) String() string {
	return fmt.Sprintf("%s\t%d\t%d\t%s\t%s\t%s", lr.MD5, lr.Timestamp, lr.Size, lr.Path, lr.MimeType, lr.MagicStr)
}

func Walk(rootPath string, out io.Writer, cache map[string]*LsRecord) ([]*LsRecord, error) {
	res := make([]*LsRecord, 0)
	fsys := os.DirFS(rootPath)
	er := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		fpath := filepath.Join(rootPath, path)
		info, err := os.Stat(fpath)
		if err != nil {
			return err
		}
		size := info.Size()
		modTime := info.ModTime().Unix()
		var lsrec *LsRecord
		if lr, ok := cache[path]; ok {
			if lr.Size == size && lr.Timestamp == modTime {
				lsrec = lr
			}
		}
		if lsrec == nil {
			magicStr := Magic(fpath)
			mtype, err := mimetype.DetectFile(fpath)
			if err != nil {
				return err
			}
			md5, err := MD5HashFunc(fpath)
			if err != nil {
				return err
			}
			lsrec = &LsRecord{MD5: md5, Timestamp: modTime, Size: size, Path: path, MimeType: mtype.String(), MagicStr: magicStr}
		}
		res = append(res, lsrec)
		fmt.Fprintf(out, "%s\n", lsrec)
		return nil
	})
	return res, er
}
