package templates

import (
	"html/template"
	"math"
	"net/url"
	"strconv"
	"time"

	"strings"

	"github.com/CloudyKit/jet"
	"github.com/NyaaPantsu/nyaa/config"
	"github.com/NyaaPantsu/nyaa/models"
	"github.com/NyaaPantsu/nyaa/models/torrents"
	"github.com/NyaaPantsu/nyaa/utils/categories"
	"github.com/NyaaPantsu/nyaa/utils/filelist"
	"github.com/NyaaPantsu/nyaa/utils/format"
	"github.com/NyaaPantsu/nyaa/utils/publicSettings"
	"github.com/NyaaPantsu/nyaa/utils/torrentLanguages"
)

// FuncMap : Functions accessible in templates by {{ $.Function }}
func templateFunctions(vars jet.VarMap) jet.VarMap {
	vars.Set("getRawQuery", getRawQuery)
	vars.Set("genSearchWithOrdering", genSearchWithOrdering)
	vars.Set("genSortArrows", genSortArrows)
	vars.Set("genNav", genNav)
	vars.Set("Sukebei", config.IsSukebei)
	vars.Set("getDefaultLanguage", publicSettings.GetDefaultLanguage)
	vars.Set("FlagCode", flagCode)
	vars.Set("getAvatar", getAvatar)

	vars.Set("formatDateRFC", formatDateRFC)
	vars.Set("GetHostname", format.GetHostname)
	vars.Set("GetCategories", categories.GetSelect)
	vars.Set("GetCategory", getCategory)
	vars.Set("CategoryName", categoryName)
	vars.Set("GetTorrentLanguages", torrentLanguages.GetTorrentLanguages)
	vars.Set("LanguageName", languageName)
	vars.Set("LanguageNameFromCode", languageNameFromCode)
	vars.Set("fileSize", fileSize)
	vars.Set("DefaultUserSettings", defaultUserSettings)
	vars.Set("makeTreeViewData", makeTreeViewData)
	vars.Set("lastID", lastID)
	vars.Set("getReportDescription", getReportDescription)
	vars.Set("genUploaderLink", genUploaderLink)
	vars.Set("genActivityContent", genActivityContent)
	vars.Set("contains", contains)
	return vars
}
func getRawQuery(currentURL *url.URL) string {
	return currentURL.RawQuery
}
func genSearchWithOrdering(currentURL *url.URL, sortBy string) string {
	values := currentURL.Query()
	order := false //Default is DESC
	sort := "2"    //Default is Date (Actually ID, but Date is the same thing)

	if _, ok := values["order"]; ok {
		order, _ = strconv.ParseBool(values["order"][0])
	}
	if _, ok := values["sort"]; ok {
		sort = values["sort"][0]
	}

	if sort == sortBy {
		order = !order //Flip order by repeat-clicking
	} else {
		order = false //Default to descending when sorting by something new
	}

	values.Set("sort", sortBy)
	values.Set("order", strconv.FormatBool(order))

	u, _ := url.Parse("/search")
	u.RawQuery = values.Encode()

	return u.String()
}

func genSortArrows(currentURL *url.URL, sortBy string) template.HTML {
	values := currentURL.Query()
	leftclass := "sortarrowdim"
	rightclass := "sortarrowdim"

	order := false
	sort := "2"

	if _, ok := values["order"]; ok {
		order, _ = strconv.ParseBool(values["order"][0])
	}
	if _, ok := values["sort"]; ok {
		sort = values["sort"][0]
	}

	if sort == sortBy {
		if order {
			rightclass = ""
		} else {
			leftclass = ""
		}
	}

	arrows := "<span class=\"sortarrowleft " + leftclass + "\">▼</span><span class=\"" + rightclass + "\">▲</span>"

	return template.HTML(arrows)
}

func genNav(nav Navigation, currentURL *url.URL, pagesSelectable int) template.HTML {
	var ret = ""
	if nav.TotalItem > 0 {
		maxPages := math.Ceil(float64(nav.TotalItem) / float64(nav.MaxItemPerPage))

		if nav.CurrentPage-1 > 0 {
			url := "/" + nav.Route + "/1"
			ret = ret + "<a id=\"page-prev\" href=\"" + url + "?" + currentURL.RawQuery + "\" aria-label=\"Previous\"><li><span aria-hidden=\"true\">&laquo;</span></li></a>"
		}
		startValue := 1
		if nav.CurrentPage > pagesSelectable/2 {
			startValue = (int(math.Min((float64(nav.CurrentPage)+math.Floor(float64(pagesSelectable)/2)), maxPages)) - pagesSelectable + 1)
		}
		if startValue < 1 {
			startValue = 1
		}
		endValue := (startValue + pagesSelectable - 1)
		if endValue > int(maxPages) {
			endValue = int(maxPages)
		}
		for i := startValue; i <= endValue; i++ {
			pageNum := strconv.Itoa(i)
			url := "/" + nav.Route + "/" + pageNum
			ret = ret + "<a aria-label=\"Page " + strconv.Itoa(i) + "\" href=\"" + url + "?" + currentURL.RawQuery + "\">" + "<li"
			if i == nav.CurrentPage {
				ret = ret + " class=\"active\""
			}
			ret = ret + ">" + strconv.Itoa(i) + "</li></a>"
		}
		if nav.CurrentPage < int(maxPages) {
			url := "/" + nav.Route + "/" + strconv.Itoa(nav.CurrentPage+1)
			ret = ret + "<a id=\"page-next\" href=\"" + url + "?" + currentURL.RawQuery + "\" aria-label=\"Next\"><li><span aria-hidden=\"true\">&raquo;</span></li></a>"
		}
		itemsThisPageStart := nav.MaxItemPerPage*(nav.CurrentPage-1) + 1
		itemsThisPageEnd := nav.MaxItemPerPage * nav.CurrentPage
		if nav.TotalItem < itemsThisPageEnd {
			itemsThisPageEnd = nav.TotalItem
		}
		ret = ret + "<p>" + strconv.Itoa(itemsThisPageStart) + "-" + strconv.Itoa(itemsThisPageEnd) + "/" + strconv.Itoa(nav.TotalItem) + "</p>"
	}
	return template.HTML(ret)
}

func flagCode(languageCode string) string {
	return publicSettings.Flag(languageCode, true)
}

func getAvatar(hash string, size int) string {
	return "https://www.gravatar.com/avatar/" + hash + "?s=" + strconv.Itoa(size)
}

func formatDateRFC(t time.Time) string {
	// because time.* isn't available in templates...
	return t.Format(time.RFC3339)
}
func getCategory(category string, keepParent bool) categories.Categories {
	cats := categories.GetSelect(true, true)
	found := false
	categoryRet := categories.Categories{}
	for _, v := range cats {
		if v.ID == category+"_" {
			found = true
			if keepParent {
				categoryRet = append(categoryRet, v)
			}
		} else if len(v.ID) <= 2 && len(categoryRet) > 0 {
			break
		} else if found {
			categoryRet = append(categoryRet, v)
		}
	}
	return categoryRet
}
func categoryName(category string, subCategory string) string {
	s := category + "_" + subCategory

	if category, ok := categories.GetByID(s); ok {
		return category.Name
	}
	return ""
}
func languageName(lang publicSettings.Language, T publicSettings.TemplateTfunc) string {
	if strings.Contains(lang.Name, ",") {
		langs := strings.Split(lang.Name, ", ")
		tags := strings.Split(lang.Tag, ", ")
		for k := range langs {
			langs[k] = strings.Title(publicSettings.Translate(tags[k], string(T("language_code"))))
		}
		return strings.Join(langs, ", ")
	}
	return strings.Title(lang.Translate(T("language_code")))
}
func languageNameFromCode(languageCode string, T publicSettings.TemplateTfunc) string {
	if strings.Contains(languageCode, ",") {
		tags := strings.Split(languageCode, ", ")
		var langs []string
		for k := range tags {
			langs = append(langs, strings.Title(publicSettings.Translate(tags[k], string(T("language_code")))))
		}
		return strings.Join(langs, ", ")
	}
	return strings.Title(publicSettings.Translate(languageCode, string(T("language_code"))))
}
func fileSize(filesize int64, T publicSettings.TemplateTfunc) template.HTML {
	if filesize == 0 {
		return T("unknown")
	}
	return template.HTML(format.FileSize(filesize))
}

func defaultUserSettings(s string) bool {
	return config.Get().Users.DefaultUserSettings[s]
}

func makeTreeViewData(f *filelist.FileListFolder, nestLevel int, identifierChain string) interface{} {
	return struct {
		Folder          *filelist.FileListFolder
		NestLevel       int
		IdentifierChain string
	}{f, nestLevel, identifierChain}
}
func lastID(currentURL *url.URL, torrents []models.TorrentJSON) int {
	if len(torrents) == 0 {
		return 0
	}
	values := currentURL.Query()

	order := false
	sort := "2"

	if _, ok := values["order"]; ok {
		order, _ = strconv.ParseBool(values["order"][0])
	}
	if _, ok := values["sort"]; ok {
		sort = values["sort"][0]
	}
	lastID := 0
	if sort == "2" || sort == "" {
		if order {
			lastID = int(torrents[len(torrents)-1].ID)
		} else if len(torrents) > 0 {
			lastID = int(torrents[0].ID)
		}
	}
	return lastID
}
func getReportDescription(d string, T publicSettings.TemplateTfunc) string {
	if d == "illegal" {
		return "Illegal content"
	} else if d == "spam" {
		return "Spam / Garbage"
	} else if d == "wrongcat" {
		return "Wrong category"
	} else if d == "dup" {
		return "Duplicate / Deprecated"
	}
	return string(T(d))
}
func genUploaderLink(uploaderID uint, uploaderName template.HTML, torrentHidden bool) template.HTML {
	uploaderID, username := torrents.HideUser(uploaderID, string(uploaderName), torrentHidden)
	if uploaderID == 0 {
		return template.HTML(username)
	}
	url := "/user/" + strconv.Itoa(int(uploaderID)) + "/" + username

	return template.HTML("<a href=\"" + url + "\">" + username + "</a>")
}
func genActivityContent(a models.Activity, T publicSettings.TemplateTfunc) template.HTML {
	return a.ToLocale(T)
}
func contains(arr interface{}, comp string) bool {
	switch str := arr.(type) {
	case string:
		if str == comp {
			return true
		}
	case publicSettings.Language:
		if str.Code == comp {
			return true
		}
	default:
		return false
	}
	return false
}
