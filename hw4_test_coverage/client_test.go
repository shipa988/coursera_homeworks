package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"
)

var xmlusers=make([]User,0,10)
var returnusers []User
var iscorruptjson bool
var is500 bool
var iscorrupterrrjson bool
var isuknownerrrjson bool
const accesstoken ="1234567890"
var timeout time.Duration
func DecodeXML(decoder *xml.Decoder) error{
	user :=User{}
	var fn string
	var ln string
	for {
		t, err := decoder.Token()
		if err != nil  {
			break
		}
		switch et := t.(type) {
		case xml.StartElement:{
						switch et.Name.Local {
						case "row":
							user =User{}
						case "id":
							decoder.DecodeElement(&user.Id, &et)
						case "age":
							decoder.DecodeElement(&user.Age, &et)
						case "first_name":
							decoder.DecodeElement(&fn, &et)
						case "last_name":
							decoder.DecodeElement(&ln, &et)
						case "gender":
							decoder.DecodeElement(&user.Gender, &et)
						case "about":
							decoder.DecodeElement(&user.About, &et)
						}
		}
		case xml.EndElement:
						if et.Name.Local == "row"{
							user.Name=fn+" "+ln
							xmlusers=append(xmlusers, user)
						}
					}
		}
	return nil
}

var CheckoutDummy = func(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("AccessToken")!=accesstoken{
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if is500{
		w.WriteHeader(http.StatusInternalServerError)
	}
	switch r.Method {
	case "GET":
		time.Sleep(time.Second*timeout)
		of:=r.URL.Query().Get("order_field")
		if  of=="Id" || of=="Age" || of=="Name"{

		} else {
			w.WriteHeader(http.StatusBadRequest)
			errorj,_:=json.Marshal(SearchErrorResponse{
				Error: "ErrorBadOrderField",
			})
			if iscorrupterrrjson {
				errorj=[]byte("{]")
			}
			if isuknownerrrjson {
				errorj,_=json.Marshal(SearchErrorResponse{
					Error: "uknownerrrjson",
				})
			}
			w.Write(errorj)
			return
		}
			resulbuf,err:=json.Marshal(returnusers)
		if err!=nil{
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var lenb=len(resulbuf)
		if iscorruptjson==true{
			lenb=lenb-5
		}
		w.Header().Add("Content-Length", strconv.Itoa(len(resulbuf)))
		w.Write(resulbuf[0:lenb])

	default:
	}

}
type testCase struct {
	returnUsersCount int
	SearchResponse
	SearchRequest
	IsError bool
	IsBadURL bool
	Timeout int
	IsCurruptJSON bool
	Iscorrupterrrjson bool
	Isuknownerrrjson bool
	AccessToken string
	Is500 bool
}

func TestFindUsers(t *testing.T) {
	xmlFile, err := os.Open("dataset.xml")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer xmlFile.Close()
	decoder := xml.NewDecoder(xmlFile)
	DecodeXML(decoder)
	testCases:=[]testCase{
		{
			Iscorrupterrrjson:true,
			AccessToken:accesstoken,
			SearchResponse:SearchResponse{
				Users:    xmlusers[:],
				NextPage: false,
			},
			SearchRequest:  SearchRequest{
				Limit:      0,
				Offset:     0,
				Query:      "",
				OrderField: "",
				OrderBy:    0,
			},
			IsError:        true,
		},
		{
			Isuknownerrrjson:true,
			AccessToken:accesstoken,
			SearchResponse:SearchResponse{
				Users:    xmlusers[:],
				NextPage: false,
			},
			SearchRequest:  SearchRequest{
				Limit:      0,
				Offset:     0,
				Query:      "",
				OrderField: "",
				OrderBy:    0,
			},
			IsError:        true,
		},
		{
			AccessToken:accesstoken,
			SearchResponse:SearchResponse{
				Users:    xmlusers[:],
				NextPage: false,
			},
			SearchRequest:  SearchRequest{
				Limit:      0,
				Offset:     0,
				Query:      "",
				OrderField: "",
				OrderBy:    0,
			},
			IsError:        true,
		},
		{
			Is500:true,
			AccessToken:accesstoken,
			SearchResponse:SearchResponse{
				Users:    xmlusers[:],
				NextPage: false,
			},
			SearchRequest:  SearchRequest{
				Limit:      0,
				Offset:     0,
				Query:      "",
				OrderField: "Id",
				OrderBy:    0,
			},
			IsError:        true,
		},
		{
			AccessToken:"bad token",
			SearchResponse:SearchResponse{
				Users:    xmlusers[:],
				NextPage: false,
			},
			SearchRequest:  SearchRequest{
				Limit:      0,
				Offset:     0,
				Query:      "",
				OrderField: "Id",
				OrderBy:    0,
			},
			IsError:        true,
		},
		{
			AccessToken:accesstoken,
			IsCurruptJSON:true,
			SearchResponse:SearchResponse{
				Users:    xmlusers[:],
				NextPage: false,
			},
			SearchRequest:  SearchRequest{
				Limit:      0,
				Offset:     0,
				Query:      "",
				OrderField: "Id",
				OrderBy:    0,
			},
			IsError:        true,
		},
		{
			AccessToken:accesstoken,
			IsBadURL:true,
			SearchResponse:SearchResponse{
				Users:    xmlusers[:],
				NextPage: false,
			},
			SearchRequest:  SearchRequest{
				Limit:      0,
				Offset:     0,
				Query:      "",
				OrderField: "Id",
				OrderBy:    0,
			},
			IsError:        true,
		},
		{
			AccessToken:accesstoken,
			Timeout:5,
			SearchResponse:SearchResponse{
				Users:    xmlusers[:],
				NextPage: false,
			},
			SearchRequest:  SearchRequest{
				Limit:      0,
				Offset:     0,
				Query:      "",
				OrderField: "Id",
				OrderBy:    0,
			},
			IsError:        true,
		},
		{
			AccessToken:accesstoken,
			SearchResponse:SearchResponse{
				Users:    xmlusers[:],
				NextPage: false,
			},
			SearchRequest:  SearchRequest{
				Limit:      0,
				Offset:     0,
				Query:      "",
				OrderField: "Id",
				OrderBy:    0,
			},
			IsError:        false,
		},
		{
			AccessToken:accesstoken,
			SearchResponse:SearchResponse{
				Users:    nil,
				NextPage: false,
			},
			SearchRequest:  SearchRequest{
				Limit:      -1,
				Offset:     0,
				Query:      "",
				OrderField: "Id",
				OrderBy:    0,
			},
			IsError:        true,
		},
		{
			AccessToken:accesstoken,
			returnUsersCount:26,
			SearchResponse:SearchResponse{
				Users:    xmlusers[:25],
				NextPage: true,
			},
			SearchRequest:  SearchRequest{
				Limit:     26,
				Offset:     0,
				Query:      "",
				OrderField: "Id",
				OrderBy:    0,
			},
			IsError:        false,
		},
		{
			AccessToken:accesstoken,
			SearchResponse:SearchResponse{
				Users:    nil,
				NextPage: false,
			},
			SearchRequest:  SearchRequest{
				Limit:      0,
				Offset:     -1,
				Query:      "",
				OrderField: "Id",
				OrderBy:    0,
			},
			IsError:        true,
		},

	}
	var testserver *httptest.Server
	for casenum,test:=range testCases  {
		func (){
			timeout= time.Duration(test.Timeout)
			iscorruptjson=test.IsCurruptJSON
			isuknownerrrjson=test.Isuknownerrrjson
			iscorrupterrrjson=test.Iscorrupterrrjson
			is500=test.Is500
			testserver= httptest.NewServer(http.HandlerFunc(CheckoutDummy))
			if test.IsBadURL{
				testserver.URL=""
			}
			defer testserver.Close()
			sc:=SearchClient{
				AccessToken: test.AccessToken,
				URL:         testserver.URL,
			}

			if test.returnUsersCount!=0{
				returnusers=xmlusers[0:test.returnUsersCount]
			}		else {returnusers=xmlusers[0:]}
			r,err:=sc.FindUsers(test.SearchRequest)
			if err==nil && test.IsError{
				t.Errorf("must be error in case %#v",casenum)
				return
			}
			if err!=nil && !test.IsError{
				t.Errorf("unexpected error  %#v in case %#v",err,casenum)
				return
			}
			if err==nil && len(r.Users)!=len(test.SearchResponse.Users){
				t.Errorf("dismatch expected users and returned %#v, %#v", len(r.Users),len(test.SearchResponse.Users))
			}
		}()


	}


}


