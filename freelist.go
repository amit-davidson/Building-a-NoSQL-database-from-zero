package LibraDB

// initialPage is the maximum pgnum that is used by the db for its own purposes. For now, only page 0 is used as the
// header page. It means all other page numbers can be used.
const initialPage = 0

// freelist manages the manages free and used pages.
type freelist struct {
	// maxPage holds the latest page num allocated. releasedPages holds all the ids that were released during
	// delete. New page ids are first given from the releasedPageIDs to avoid growing the file. If it's empty, then
	// maxPage is incremented and a new page is created thus increasing the file size.
	maxPage       pgnum
	releasedPages []pgnum
}

func newFreelist() *freelist {
	return &freelist{
		maxPage:       initialPage,
		releasedPages: []pgnum{},
	}
}

// getNextPage returns page ids for writing New page ids are first given from the releasedPageIDs to avoid growing
// the file. If it's empty, then maxPage is incremented and a new page is created thus increasing the file size.
func (fr *freelist) getNextPage() pgnum {
	if len(fr.releasedPages) != 0 {
		// Take the last element and remove it from the list
		pageID := fr.releasedPages[len(fr.releasedPages)-1]
		fr.releasedPages = fr.releasedPages[:len(fr.releasedPages)-1]
		return pageID
	}
	fr.maxPage += 1
	return fr.maxPage
}

func (fr *freelist) releasePage(page pgnum) {
	fr.releasedPages = append(fr.releasedPages, page)
}

