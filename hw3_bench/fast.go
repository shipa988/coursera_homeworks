package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func ReadFileAtLeast(path string) ([]byte,error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return []byte{},err
	}
	fileinfo,err:=file.Stat()
	if err != nil {
		return  []byte{},err
	}
	fileContents:=make([]byte,fileinfo.Size())
	_,err=io.ReadAtLeast(file,fileContents,len(fileContents))
	if err != nil {
		return  []byte{},err
	}
	return fileContents,nil
}
// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {

	fileContents, err := ReadFileAtLeast(filePath)
		if err != nil {
			panic(err)
	}
	bbfile:=bytes.Split(fileContents, []byte("\n"))
	isAndroid,isAndroidch,isMSIE,isMSIEch := false,false,false,false
	email:=""
	const(
		Androidstr="Android"
		MSIEstr="MSIE"
	)
	var Androidbyte=[]byte(Androidstr)
	var MSIEbyte=[]byte(MSIEstr)
	user := User{}
	foundUsers := strings.Builder{}
	foundUsers.Grow(len(bbfile))
	seenBrowsers := make(map[string]bool,5)
	for i,bline:=range bbfile   {
		if bytes.Contains(bline,Androidbyte)|| bytes.Contains(bline,MSIEbyte){
			err =user.UnmarshalJSON(bline)
			if err != nil {
				panic(err)
			}
			isAndroid,isMSIE = false,false
			for _, browser := range user.Browsers {
				isAndroidch=(strings.Contains(browser,Androidstr ))
				isMSIEch=(strings.Contains(browser,MSIEstr ))
				isAndroid=isAndroid||isAndroidch
				isMSIE=isMSIE||isMSIEch
				if isAndroidch || isMSIEch {
					if _,ok:=seenBrowsers[browser];!ok{
						seenBrowsers[browser]=true
					}
				}
			}
			if !(isAndroid && isMSIE) {
				continue
			}
			email = strings.ReplaceAll(user.Email, "@"," [at] ")
			foundUsers.WriteString(fmt.Sprintf("[%s] %s <%s>\n", strconv.Itoa(i), user.Name, email))
		}
	}
	fmt.Fprintln(out, "found users:\n"+foundUsers.String())
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}
