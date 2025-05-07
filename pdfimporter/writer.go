package pdfimporter

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"math"
	"os"

	"github.com/skirrund/go-pdf/pdfimporter/errors"
)

type PdfWriter struct {
	f       *os.File
	w       *bufio.Writer
	r       *PdfReader
	k       float64
	tpls    []*PdfTemplate
	m       int
	n       int
	offsets map[int]int
	offset  int
	result  map[int]string
	// Keep track of which objects have already been written
	obj_stack       map[int]*PdfValue
	don_obj_stack   map[int]*PdfValue
	written_objs    map[*PdfObjectId][]byte
	written_obj_pos map[*PdfObjectId]map[int]string
	current_obj     *PdfObject
	current_obj_id  int
	tpl_id_offset   int
	use_hash        bool
}

type PdfObjectId struct {
	id   int
	hash string
}

type PdfObject struct {
	id     *PdfObjectId
	buffer *bytes.Buffer
}

func (pr *PdfWriter) SetTplIdOffset(n int) {
	pr.tpl_id_offset = n
}

func (pr *PdfWriter) Init() {
	pr.k = 1
	pr.obj_stack = make(map[int]*PdfValue, 0)
	pr.don_obj_stack = make(map[int]*PdfValue, 0)
	pr.tpls = make([]*PdfTemplate, 0)
	pr.written_objs = make(map[*PdfObjectId][]byte, 0)
	pr.written_obj_pos = make(map[*PdfObjectId]map[int]string, 0)
	pr.current_obj = new(PdfObject)
}

func (pr *PdfWriter) SetUseHash(b bool) {
	pr.use_hash = b
}

func (pr *PdfWriter) SetNextObjectID(id int) {
	pr.n = id - 1
}

func NewPdfWriter(filename string) (*PdfWriter, error) {
	writer := &PdfWriter{}
	writer.Init()

	if filename != "" {
		var err error
		f, err := os.Create(filename)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to create filename: "+filename)
		}
		writer.f = f
		writer.w = bufio.NewWriter(f)
	}
	return writer, nil
}

// Done with parsing.  Now, create templates.
type PdfTemplate struct {
	Id        int
	Reader    *PdfReader
	Resources *PdfValue
	Buffer    string
	Box       map[string]float64
	Boxes     map[string]map[string]float64
	X         float64
	Y         float64
	W         float64
	H         float64
	Rotation  int
	N         int
}

func (pr *PdfWriter) GetImportedObjects() map[*PdfObjectId][]byte {
	return pr.written_objs
}

// For each object (uniquely identified by a sha1 hash), return the positions
// of each hash within the object, to be replaced with pdf object ids (integers)
func (pr *PdfWriter) GetImportedObjHashPos() map[*PdfObjectId]map[int]string {
	return pr.written_obj_pos
}

func (pr *PdfWriter) ClearImportedObjects() {
	pr.written_objs = make(map[*PdfObjectId][]byte, 0)
}

// Create a PdfTemplate object from a page number (e.g. 1) and a boxName (e.g. MediaBox)
func (pr *PdfWriter) ImportPage(reader *PdfReader, pageno int, boxName string) (int, error) {
	var err error

	// Set default scale to 1
	pr.k = 1

	// Get all page boxes
	pageBoxes, err := reader.getPageBoxes(1, pr.k)
	if err != nil {
		return -1, errors.Wrap(err, "Failed to get page boxes")
	}

	// If requested box name does not exist for this page, use an alternate box
	if _, ok := pageBoxes[boxName]; !ok {
		if boxName == "/BleedBox" || boxName == "/TrimBox" || boxName == "ArtBox" {
			boxName = "/CropBox"
		} else if boxName == "/CropBox" {
			boxName = "/MediaBox"
		}
	}

	// If the requested box name or an alternate box name cannot be found, trigger an error
	// TODO: Improve error handling
	if _, ok := pageBoxes[boxName]; !ok {
		return -1, errors.New("Box not found: " + boxName)
	}

	pageResources, err := reader.getPageResources(pageno)
	if err != nil {
		return -1, errors.Wrap(err, "Failed to get page resources")
	}

	content, err := reader.getContent(pageno)
	if err != nil {
		return -1, errors.Wrap(err, "Failed to get content")
	}

	// Set template values
	tpl := &PdfTemplate{}
	tpl.Reader = reader
	tpl.Resources = pageResources
	tpl.Buffer = content
	tpl.Box = pageBoxes[boxName]
	tpl.Boxes = pageBoxes
	tpl.X = 0
	tpl.Y = 0
	tpl.W = tpl.Box["w"]
	tpl.H = tpl.Box["h"]

	// Set template rotation
	rotation, err := reader.getPageRotation(pageno)
	if err != nil {
		return -1, errors.Wrap(err, "Failed to get page rotation")
	}
	angle := rotation.Int % 360

	// Normalize angle
	if angle != 0 {
		steps := angle / 90
		w := tpl.W
		h := tpl.H

		if steps%2 == 0 {
			tpl.W = w
			tpl.H = h
		} else {
			tpl.W = h
			tpl.H = w
		}

		if angle < 0 {
			angle += 360
		}

		tpl.Rotation = angle * -1
	}

	pr.tpls = append(pr.tpls, tpl)

	// Return last template id
	return len(pr.tpls) - 1, nil
}

// Create a new object and keep track of the offset for the xref table
func (pr *PdfWriter) newObj(objId int, onlyNewObj bool) {
	if objId < 0 {
		pr.n++
		objId = pr.n
	}

	if !onlyNewObj {
		// set current object id integer
		pr.current_obj_id = objId

		// Create new PdfObject and PdfObjectId
		pr.current_obj = new(PdfObject)
		pr.current_obj.buffer = new(bytes.Buffer)
		pr.current_obj.id = new(PdfObjectId)
		pr.current_obj.id.id = objId
		pr.current_obj.id.hash = pr.shaOfInt(objId)

		pr.written_obj_pos[pr.current_obj.id] = make(map[int]string, 0)
	}
}

func (pr *PdfWriter) endObj() {
	pr.out("endobj")

	pr.written_objs[pr.current_obj.id] = pr.current_obj.buffer.Bytes()
	pr.current_obj_id = -1
}

func (pr *PdfWriter) shaOfInt(i int) string {
	hasher := sha1.New()
	hasher.Write([]byte(fmt.Sprintf("%d-%s", i, pr.r.sourceFile)))
	sha := hex.EncodeToString(hasher.Sum(nil))
	return sha
}

func (pr *PdfWriter) outObjRef(objId int) {
	sha := pr.shaOfInt(objId)

	// Keep track of object hash and position - to be replaced with actual object id (integer)
	pr.written_obj_pos[pr.current_obj.id][pr.current_obj.buffer.Len()] = sha

	if pr.use_hash {
		pr.current_obj.buffer.WriteString(sha)
	} else {
		pr.current_obj.buffer.WriteString(fmt.Sprintf("%d", objId))
	}
	pr.current_obj.buffer.WriteString(" 0 R ")
}

// Output PDF data with a newline
func (pr *PdfWriter) out(s string) {
	pr.current_obj.buffer.WriteString(s)
	pr.current_obj.buffer.WriteString("\n")
}

// Output PDF data
func (pr *PdfWriter) straightOut(s string) {
	pr.current_obj.buffer.WriteString(s)
}

// Output a PdfValue
func (pr *PdfWriter) writeValue(value *PdfValue) {
	switch value.Type {
	case PDF_TYPE_TOKEN:
		pr.straightOut(value.Token + " ")

	case PDF_TYPE_NUMERIC:
		pr.straightOut(fmt.Sprintf("%d", value.Int) + " ")

	case PDF_TYPE_REAL:
		pr.straightOut(fmt.Sprintf("%F", value.Real) + " ")

	case PDF_TYPE_ARRAY:
		pr.straightOut("[")
		for i := 0; i < len(value.Array); i++ {
			pr.writeValue(value.Array[i])
		}
		pr.out("]")

	case PDF_TYPE_DICTIONARY:
		pr.straightOut("<<")
		for k, v := range value.Dictionary {
			pr.straightOut(k + " ")
			pr.writeValue(v)
		}
		pr.straightOut(">>")

	case PDF_TYPE_OBJREF:
		// An indirect object reference.  Fill the object stack if needed.
		// Check to see if object already exists on the don_obj_stack.
		if _, ok := pr.don_obj_stack[value.Id]; !ok {
			pr.newObj(-1, true)
			pr.obj_stack[value.Id] = &PdfValue{Type: PDF_TYPE_OBJREF, Gen: value.Gen, Id: value.Id, NewId: pr.n}
			pr.don_obj_stack[value.Id] = &PdfValue{Type: PDF_TYPE_OBJREF, Gen: value.Gen, Id: value.Id, NewId: pr.n}
		}

		// Get object ID from don_obj_stack
		objId := pr.don_obj_stack[value.Id].NewId
		pr.outObjRef(objId)
		//this.out(fmt.Sprintf("%d 0 R", objId))

	case PDF_TYPE_STRING:
		// A string
		pr.straightOut("(" + value.String + ")")

	case PDF_TYPE_STREAM:
		// A stream.  First, output the stream dictionary, then the stream data itself.
		pr.writeValue(value.Value)
		pr.out("stream")
		pr.out(string(value.Stream.Bytes))
		pr.out("endstream")

	case PDF_TYPE_HEX:
		pr.straightOut("<" + value.String + ">")

	case PDF_TYPE_BOOLEAN:
		if value.Bool {
			pr.straightOut("true ")
		} else {
			pr.straightOut("false ")
		}

	case PDF_TYPE_NULL:
		// The null object
		pr.straightOut("null ")
	}
}

// Output Form XObjects (1 for each template)
// returns a map of template names (e.g. /GOFPDITPL1) to PdfObjectId
func (pr *PdfWriter) PutFormXobjects(reader *PdfReader) (map[string]*PdfObjectId, error) {
	// Set current reader
	pr.r = reader

	var err error
	var result = make(map[string]*PdfObjectId, 0)

	compress := true
	filter := ""
	if compress {
		filter = "/Filter /FlateDecode "
	}

	for i := 0; i < len(pr.tpls); i++ {
		tpl := pr.tpls[i]
		if tpl == nil {
			return nil, errors.New("Template is nil")
		}
		var p string
		if compress {
			var b bytes.Buffer
			w := zlib.NewWriter(&b)
			w.Write([]byte(tpl.Buffer))
			w.Close()

			p = b.String()
		} else {
			p = tpl.Buffer
		}

		// Create new PDF object
		pr.newObj(-1, false)

		cN := pr.n // remember current "n"

		tpl.N = pr.n

		// Return xobject form name and object position
		pdfObjId := new(PdfObjectId)
		pdfObjId.id = cN
		pdfObjId.hash = pr.shaOfInt(cN)
		result[fmt.Sprintf("/GOFPDITPL%d", i+pr.tpl_id_offset)] = pdfObjId

		pr.out("<<" + filter + "/Type /XObject")
		pr.out("/Subtype /Form")
		pr.out("/FormType 1")

		pr.out(fmt.Sprintf("/BBox [%.2F %.2F %.2F %.2F]", tpl.Box["llx"]*pr.k, tpl.Box["lly"]*pr.k, (tpl.Box["urx"]+tpl.X)*pr.k, (tpl.Box["ury"]-tpl.Y)*pr.k))

		var c, s, tx, ty float64
		c = 1

		// Handle rotated pages
		if tpl.Box != nil {
			tx = -tpl.Box["llx"]
			ty = -tpl.Box["lly"]

			if tpl.Rotation != 0 {
				angle := float64(tpl.Rotation) * math.Pi / 180.0
				c = math.Cos(float64(angle))
				s = math.Sin(float64(angle))

				switch tpl.Rotation {
				case -90:
					tx = -tpl.Box["lly"]
					ty = tpl.Box["urx"]

				case -180:
					tx = tpl.Box["urx"]
					ty = tpl.Box["ury"]

				case -270:
					tx = tpl.Box["ury"]
					ty = -tpl.Box["llx"]
				}
			}
		} else {
			tx = -tpl.Box["x"] * 2
			ty = tpl.Box["y"] * 2
		}

		tx *= pr.k
		ty *= pr.k

		if c != 1 || s != 0 || tx != 0 || ty != 0 {
			pr.out(fmt.Sprintf("/Matrix [%.5F %.5F %.5F %.5F %.5F %.5F]", c, s, -s, c, tx, ty))
		}

		// Now write resources
		pr.out("/Resources ")

		if tpl.Resources != nil {
			pr.writeValue(tpl.Resources) // "n" will be changed
		} else {
			return nil, errors.New("Template resources are empty")
		}

		nN := pr.n // remember new "n"
		pr.n = cN  // reset to current "n"

		pr.out("/Length " + fmt.Sprintf("%d", len(p)) + " >>")

		pr.out("stream")
		pr.out(p)
		pr.out("endstream")

		pr.endObj()

		pr.n = nN // reset to new "n"

		// Put imported objects, starting with the ones from the XObject's Resources,
		// then from dependencies of those resources).
		err = pr.putImportedObjects(reader)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to put imported objects")
		}
	}

	return result, nil
}

func (pr *PdfWriter) putImportedObjects(reader *PdfReader) error {
	var err error
	var nObj *PdfValue

	// obj_stack will have new items added to it in the inner loop, so do another loop to check for extras
	// TODO make the order of this the same every time
	for {
		atLeastOne := false

		// FIXME:  How to determine number of objects before this loop?
		for i := 0; i < 9999; i++ {
			k := i
			v := pr.obj_stack[i]

			if v == nil {
				continue
			}

			atLeastOne = true

			nObj, err = reader.resolveObject(v)
			if err != nil {
				return errors.Wrap(err, "Unable to resolve object")
			}

			// New object with "NewId" field
			pr.newObj(v.NewId, false)

			if nObj.Type == PDF_TYPE_STREAM {
				pr.writeValue(nObj)
			} else {
				pr.writeValue(nObj.Value)
			}

			pr.endObj()

			// Remove from stack
			pr.obj_stack[k] = nil
		}

		if !atLeastOne {
			break
		}
	}

	return nil
}

// Get the calculated size of a template
// If one size is given, this method calculates the other one
func (pr *PdfWriter) getTemplateSize(tplid int, _w float64, _h float64) map[string]float64 {
	result := make(map[string]float64, 2)

	tpl := pr.tpls[tplid]

	w := tpl.W
	h := tpl.H

	if _w == 0 && _h == 0 {
		_w = w
		_h = h
	}

	if _w == 0 {
		_w = _h * w / h
	}

	if _h == 0 {
		_h = _w * h / w
	}

	result["w"] = _w
	result["h"] = _h

	return result
}

func (pr *PdfWriter) UseTemplate(tplid int, _x float64, _y float64, _w float64, _h float64) (string, float64, float64, float64, float64) {
	tpl := pr.tpls[tplid]

	w := tpl.W
	h := tpl.H

	_x += tpl.X
	_y += tpl.Y

	wh := pr.getTemplateSize(0, _w, _h)

	_w = wh["w"]
	_h = wh["h"]

	tData := make(map[string]float64, 9)
	tData["x"] = 0.0
	tData["y"] = 0.0
	tData["w"] = _w
	tData["h"] = _h
	tData["scaleX"] = (_w / w)
	tData["scaleY"] = (_h / h)
	tData["tx"] = _x
	tData["ty"] = (0 - _y - _h)
	tData["lty"] = (0 - _y - _h) - (0-h)*(_h/h)

	return fmt.Sprintf("/GOFPDITPL%d", tplid+pr.tpl_id_offset), tData["scaleX"], tData["scaleY"], tData["tx"] * pr.k, tData["ty"] * pr.k
}
