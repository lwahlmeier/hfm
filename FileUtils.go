package main

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

type VirtualDirs struct {
	VirtualDirs map[string][]string `yaml:"virtualdirs"`
}

type MergedFileSystem struct {
	paths   map[string][]string
	dirs    map[string][]http.Dir
	fakeDir *FakeDir
}

func NewMergedFileSystem(dirMap map[string][]string) *MergedFileSystem {
	tdirs := make(map[string][]http.Dir)
	tpaths := make(map[string][]string)
	lfd := make([]os.FileInfo, 0)
	for k, v := range dirMap {
		sl := make([]string, len(v))
		dl := make([]http.Dir, len(v))
		for i, v2 := range v {
			sl[i] = v2
			dl[i] = http.Dir(v2)
		}
		tdirs[k] = dl
		tpaths[k] = sl
		fdfi := &FakeDirFileInfo{
			name: k,
		}
		lfd = append(lfd, fdfi)
	}

	fd := &FakeDir{
		name: "",
		fis:  lfd,
	}
	return &MergedFileSystem{paths: tpaths, dirs: tdirs, fakeDir: fd}
}

func (mfs *MergedFileSystem) Open(name string) (File, error) {
	if name == "" || name == "/" {
		return mfs.fakeDir, nil
	}
	nname := path.Clean("/" + name)

	if name[0] == os.PathSeparator {
		nname = name[1:]
	}
	spath := strings.Split(nname, string(os.PathSeparator))

	if dirs, ok := mfs.dirs[spath[0]]; ok {

		np := string(os.PathSeparator) + strings.Join(spath[1:], "/")
		fpl := make([]http.File, 0)
		for _, bf := range dirs {
			fp, err := bf.Open("/" + np)
			if err != nil {
				continue
			}
			fpl = append(fpl, fp)
		}
		if len(fpl) == 0 {
			return nil, errors.New("can not open path:" + name)
		}
		return &MergedFile{
			pathFile:   fpl,
			hideHidden: !config.GetBool(SHOWHIDDEN),
		}, nil
	}
	return nil, errors.New("can not open path:" + name)
}

type MergedFile struct {
	pathFile   []http.File
	hideHidden bool
}

func (mf *MergedFile) Close() error {
	return mf.pathFile[0].Close()
}

func (mf *MergedFile) Read(p []byte) (n int, err error) {
	return mf.pathFile[0].Read(p)
}

func (mf *MergedFile) Seek(offset int64, whence int) (int64, error) {
	return mf.pathFile[0].Seek(offset, whence)
}

func (mf *MergedFile) Readdir(count int) ([]os.FileInfo, error) {

	if len(mf.pathFile) > 1 {
		var lastErr error
		fim := make(map[string]os.FileInfo)
		for _, fp := range mf.pathFile {
			tfil, err := fp.Readdir(count)
			fp.Close()
			if err != nil {
				lastErr = err
				continue
			}
			for _, fi := range tfil {
				if _, ok := fim[fi.Name()]; !ok {
					fim[fi.Name()] = fi
				}
				if count > 0 && len(fim) >= count {
					break
				}
			}
		}
		if len(fim) == 0 {
			return nil, lastErr
		}
		fil := make([]os.FileInfo, 0)
		for _, v := range fim {
			if mf.hideHidden && v.Name()[0] == '.' {
				continue
			}
			fil = append(fil, v)
		}
		return fil, nil
	}
	if mf.hideHidden {
		ofil, err := mf.pathFile[0].Readdir(count)
		mf.pathFile[0].Close()
		fil := make([]os.FileInfo, 0)
		for _, v := range ofil {
			if v.Name()[0] == '.' {
				continue
			}
			fil = append(fil, v)
		}
		return fil, err
	}
	return mf.pathFile[0].Readdir(count)
}

func (mf *MergedFile) Stat() (os.FileInfo, error) {
	return mf.pathFile[0].Stat()
}

type FakeDir struct {
	name string
	fis  []os.FileInfo
	fi   *FakeDirFileInfo
}

func (fd *FakeDir) Close() error {
	return errors.New("This is a Directory")
}

func (fd *FakeDir) Read(p []byte) (n int, err error) {
	return 0, errors.New("This is a Directory")
}

func (fd *FakeDir) Seek(offset int64, whence int) (int64, error) {
	return 0, errors.New("This is a Directory")
}

func (fd *FakeDir) Readdir(count int) ([]os.FileInfo, error) {
	if count < 0 || count >= len(fd.fis) {
		return fd.fis[:], nil
	}
	return fd.fis[:count], nil
}

func (fd *FakeDir) Stat() (os.FileInfo, error) {
	if fd.fi == nil {
		fd.fi = &FakeDirFileInfo{name: fd.name}
	}
	return fd.fi, nil
}

type FakeDirFileInfo struct {
	name string
}

func (fdfi *FakeDirFileInfo) Name() string {
	return fdfi.name
}
func (fdfi *FakeDirFileInfo) Size() int64 {
	return 0
}
func (fdfi *FakeDirFileInfo) Mode() os.FileMode {
	return os.ModeDir | 0777
}
func (fdfi *FakeDirFileInfo) ModTime() time.Time {
	return time.Now()
}
func (fdfi *FakeDirFileInfo) IsDir() bool {
	return true
}
func (fdfi *FakeDirFileInfo) Sys() interface{} {
	return nil
}

var BufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 8*1024)
	},
}

// copyBuffer is the actual implementation of Copy and CopyBuffer.
// if buf is nil, one is allocated.
func copyBufferN(dst io.Writer, src io.Reader, nsize int64) (written int64, err error) {
	buf := BufferPool.Get().([]byte)
	defer BufferPool.Put(buf)
	rl := io.LimitReader(src, nsize)
	for {
		nr, er := rl.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}

	return written, err
}
