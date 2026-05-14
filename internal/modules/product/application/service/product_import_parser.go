package service

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Taishi-co/SteelLink-backend/config"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/product/ui/dto"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
)

const maxProductImportFileBytes = 10 << 20

type productImportSource struct {
	Filename    string
	SourceType  string
	Text        string
	DataURL     string
	IsImage     bool
	ContentType string
	Warnings    []string
}

func readProductImportSource(header *multipart.FileHeader) (productImportSource, error) {
	if header == nil {
		return productImportSource{}, ecode.ErrBadRequest.WithMessage("请选择要导入的台账或截图")
	}
	if header.Size <= 0 {
		return productImportSource{}, ecode.ErrBadRequest.WithMessage("导入文件为空")
	}
	if header.Size > maxProductImportFileBytes {
		return productImportSource{}, ecode.ErrBadRequest.WithMessage("导入文件不能超过 10MB")
	}

	file, err := header.Open()
	if err != nil {
		return productImportSource{}, ecode.ErrBadRequest.WithMessage("读取导入文件失败")
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxProductImportFileBytes+1))
	if err != nil {
		return productImportSource{}, ecode.ErrBadRequest.WithMessage("读取导入文件失败")
	}
	if len(data) > maxProductImportFileBytes {
		return productImportSource{}, ecode.ErrBadRequest.WithMessage("导入文件不能超过 10MB")
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	source := productImportSource{
		Filename:    header.Filename,
		ContentType: header.Header.Get("Content-Type"),
	}
	switch ext {
	case ".csv":
		rows := parseDelimitedImportRows(string(data))
		source.SourceType = "csv"
		source.Text = renderImportRows(rows)
	case ".tsv":
		rows := parseImportRowsWithDelimiter(string(data), '\t')
		source.SourceType = "tsv"
		source.Text = renderImportRows(rows)
	case ".txt":
		source.SourceType = "text"
		source.Text = string(data)
	case ".xlsx":
		rows, err := readXLSXImportRows(data)
		if err != nil {
			return productImportSource{}, ecode.ErrBadRequest.WithMessage("解析 xlsx 台账失败: " + err.Error())
		}
		source.SourceType = "xlsx"
		source.Text = renderImportRows(rows)
	case ".jpg", ".jpeg", ".png", ".webp":
		contentType := source.ContentType
		if contentType == "" {
			switch ext {
			case ".jpg", ".jpeg":
				contentType = "image/jpeg"
			case ".png":
				contentType = "image/png"
			case ".webp":
				contentType = "image/webp"
			}
		}
		source.SourceType = "image"
		source.IsImage = true
		source.DataURL = fmt.Sprintf("data:%s;base64,%s", contentType, base64.StdEncoding.EncodeToString(data))
	case ".xls":
		return productImportSource{}, ecode.ErrBadRequest.WithMessage("暂不支持旧版 xls，请另存为 xlsx 或 csv 后导入")
	default:
		return productImportSource{}, ecode.ErrBadRequest.WithMessage("仅支持 csv、tsv、txt、xlsx、jpg、jpeg、png、webp")
	}

	source.Text = strings.TrimSpace(source.Text)
	return source, nil
}

func parseDelimitedImportRows(text string) [][]string {
	delimiters := []rune{',', '\t', ';', '，'}
	var best [][]string
	bestScore := -1
	for _, delimiter := range delimiters {
		rows := parseImportRowsWithDelimiter(text, delimiter)
		score := scoreDelimitedRows(rows)
		if score > bestScore {
			best = rows
			bestScore = score
		}
	}
	return best
}

func parseImportRowsWithDelimiter(text string, delimiter rune) [][]string {
	reader := csv.NewReader(strings.NewReader(text))
	reader.Comma = delimiter
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true
	reader.LazyQuotes = true
	rows, err := reader.ReadAll()
	if err != nil {
		return splitImportTextRows(text)
	}
	return cleanImportRows(rows)
}

func scoreDelimitedRows(rows [][]string) int {
	score := 0
	for _, row := range rows {
		if len(row) > 1 {
			score += len(row)
		}
	}
	return score
}

func splitImportTextRows(text string) [][]string {
	lines := strings.Split(text, "\n")
	rows := make([][]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		rows = append(rows, regexp.MustCompile(`[\t,，;|]+`).Split(line, -1))
	}
	return cleanImportRows(rows)
}

func cleanImportRows(rows [][]string) [][]string {
	result := make([][]string, 0, len(rows))
	for _, row := range rows {
		cleaned := make([]string, 0, len(row))
		hasValue := false
		for _, cell := range row {
			cell = strings.TrimSpace(strings.TrimPrefix(cell, "\ufeff"))
			cleaned = append(cleaned, cell)
			if cell != "" {
				hasValue = true
			}
		}
		if hasValue {
			result = append(result, cleaned)
		}
	}
	return result
}

func renderImportRows(rows [][]string) string {
	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		lines = append(lines, strings.Join(row, "\t"))
	}
	return strings.Join(lines, "\n")
}

type xlsxSharedStrings struct {
	Items []xlsxSharedStringItem `xml:"si"`
}

type xlsxSharedStringItem struct {
	TextRuns []string `xml:"r>t"`
	Text     string   `xml:"t"`
}

type xlsxWorksheet struct {
	Rows []xlsxRow `xml:"sheetData>row"`
}

type xlsxRow struct {
	Cells []xlsxCell `xml:"c"`
}

type xlsxCell struct {
	Ref        string `xml:"r,attr"`
	Type       string `xml:"t,attr"`
	Value      string `xml:"v"`
	InlineText string `xml:"is>t"`
}

func readXLSXImportRows(data []byte) ([][]string, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}

	sharedStrings := map[int]string{}
	for _, file := range reader.File {
		if file.Name != "xl/sharedStrings.xml" {
			continue
		}
		content, err := readZipFile(file)
		if err != nil {
			return nil, err
		}
		var parsed xlsxSharedStrings
		if err := xml.Unmarshal(content, &parsed); err != nil {
			return nil, err
		}
		for index, item := range parsed.Items {
			value := item.Text
			if value == "" && len(item.TextRuns) > 0 {
				value = strings.Join(item.TextRuns, "")
			}
			sharedStrings[index] = strings.TrimSpace(value)
		}
	}

	sheetNames := make([]string, 0)
	for _, file := range reader.File {
		if strings.HasPrefix(file.Name, "xl/worksheets/sheet") && strings.HasSuffix(file.Name, ".xml") {
			sheetNames = append(sheetNames, file.Name)
		}
	}
	sort.Strings(sheetNames)
	if len(sheetNames) == 0 {
		return nil, fmt.Errorf("未找到工作表")
	}

	for _, sheetName := range sheetNames {
		file := findZipFile(reader, sheetName)
		if file == nil {
			continue
		}
		content, err := readZipFile(file)
		if err != nil {
			return nil, err
		}
		var worksheet xlsxWorksheet
		if err := xml.Unmarshal(content, &worksheet); err != nil {
			return nil, err
		}
		rows := xlsxRowsToStrings(worksheet.Rows, sharedStrings)
		if len(rows) > 0 {
			return rows, nil
		}
	}
	return nil, fmt.Errorf("工作表为空")
}

func readZipFile(file *zip.File) ([]byte, error) {
	rc, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

func findZipFile(reader *zip.Reader, name string) *zip.File {
	for _, file := range reader.File {
		if file.Name == name {
			return file
		}
	}
	return nil
}

func xlsxRowsToStrings(rows []xlsxRow, sharedStrings map[int]string) [][]string {
	result := make([][]string, 0, len(rows))
	for _, row := range rows {
		values := make([]string, 0, len(row.Cells))
		for _, cell := range row.Cells {
			index := xlsxColumnIndex(cell.Ref)
			for index >= len(values) {
				values = append(values, "")
			}
			values[index] = xlsxCellValue(cell, sharedStrings)
		}
		result = append(result, values)
	}
	return cleanImportRows(result)
}

func xlsxColumnIndex(ref string) int {
	index := 0
	hasColumn := false
	for _, ch := range ref {
		if ch >= 'A' && ch <= 'Z' {
			index = index*26 + int(ch-'A'+1)
			hasColumn = true
			continue
		}
		if ch >= 'a' && ch <= 'z' {
			index = index*26 + int(ch-'a'+1)
			hasColumn = true
			continue
		}
		break
	}
	if !hasColumn {
		return 0
	}
	return index - 1
}

func xlsxCellValue(cell xlsxCell, sharedStrings map[int]string) string {
	switch cell.Type {
	case "s":
		index, _ := strconv.Atoi(strings.TrimSpace(cell.Value))
		return sharedStrings[index]
	case "inlineStr":
		return strings.TrimSpace(cell.InlineText)
	default:
		return strings.TrimSpace(cell.Value)
	}
}

func parseProductImportRows(text string) []dto.ProductImportDraftItem {
	rows := parseDelimitedImportRows(text)
	if len(rows) == 0 {
		rows = splitImportTextRows(text)
	}
	items := parseProductImportTableRows(rows, 0.58)
	if len(items) > 0 {
		return items
	}

	lines := strings.Split(text, "\n")
	items = make([]dto.ProductImportDraftItem, 0, len(lines))
	for _, line := range lines {
		item, ok := inferProductImportLine(line)
		if ok {
			items = append(items, item)
		}
	}
	return normalizeImportDraftItems(items)
}

func parseProductImportTableRows(rows [][]string, confidence float64) []dto.ProductImportDraftItem {
	rows = cleanImportRows(rows)
	if len(rows) == 0 {
		return nil
	}

	headerIndex, headerMap := detectProductImportHeader(rows)
	if headerIndex < 0 {
		items := make([]dto.ProductImportDraftItem, 0, len(rows))
		for _, row := range rows {
			item, ok := inferProductImportLine(strings.Join(row, " "))
			if ok {
				item.Confidence = confidence
				items = append(items, item)
			}
		}
		return normalizeImportDraftItems(items)
	}

	items := make([]dto.ProductImportDraftItem, 0, len(rows)-headerIndex-1)
	for _, row := range rows[headerIndex+1:] {
		if isImportRowEmpty(row) {
			continue
		}
		product := dto.ProductImportProduct{
			Name:        importCell(row, headerMap, "name"),
			GoodsType:   normalizeImportGoodsType(importCell(row, headerMap, "goods_type")),
			Category:    importCell(row, headerMap, "category"),
			SubCategory: importCell(row, headerMap, "sub_category"),
			SteelMill:   importCell(row, headerMap, "steel_mill"),
			BundleNo:    importCell(row, headerMap, "bundle_no"),
			Quantity:    parseImportNumber(importCell(row, headerMap, "quantity")),
			UnitPrice:   parseImportNumber(importCell(row, headerMap, "unit_price")),
			MinOrderQty: parseImportNumber(importCell(row, headerMap, "min_order_qty")),
			Description: importCell(row, headerMap, "description"),
			Spec: dto.ProductSpec{
				Material: importCell(row, headerMap, "material"),
				Spec:     importCell(row, headerMap, "spec"),
			},
			WarehouseInfo: dto.WarehouseInfo{
				WarehouseName: importCell(row, headerMap, "warehouse_name"),
				WarehouseCode: importCell(row, headerMap, "warehouse_code"),
				Province:      importCell(row, headerMap, "province"),
				City:          importCell(row, headerMap, "city"),
				District:      importCell(row, headerMap, "district"),
				Address:       importCell(row, headerMap, "address"),
			},
			DeliveryDays:  3,
			QuoteValidity: 72,
		}
		items = append(items, dto.ProductImportDraftItem{Confidence: confidence, Product: product})
	}
	return normalizeImportDraftItems(items)
}

func detectProductImportHeader(rows [][]string) (int, map[string]int) {
	limit := len(rows)
	if limit > 8 {
		limit = 8
	}
	for i := 0; i < limit; i++ {
		headerMap := map[string]int{}
		score := 0
		for idx, cell := range rows[i] {
			kind := productImportHeaderKind(cell)
			if kind == "" {
				continue
			}
			if _, exists := headerMap[kind]; !exists {
				headerMap[kind] = idx
				score++
			}
		}
		if score >= 2 {
			return i, headerMap
		}
	}
	return -1, nil
}

func productImportHeaderKind(value string) string {
	value = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(value), " ", ""))
	switch {
	case value == "":
		return ""
	case strings.Contains(value, "商品名称") || strings.Contains(value, "品名") || value == "name":
		return "name"
	case strings.Contains(value, "货品类型") || strings.Contains(value, "现货") || strings.Contains(value, "期货"):
		return "goods_type"
	case strings.Contains(value, "细分品种") || strings.Contains(value, "子类"):
		return "sub_category"
	case strings.Contains(value, "品种") || strings.Contains(value, "分类") || strings.Contains(value, "category"):
		return "category"
	case strings.Contains(value, "牌号") || strings.Contains(value, "材质") || strings.Contains(value, "material"):
		return "material"
	case strings.Contains(value, "规格") || strings.Contains(value, "尺寸") || strings.Contains(value, "spec"):
		return "spec"
	case strings.Contains(value, "产地") || strings.Contains(value, "钢厂") || strings.Contains(value, "厂家"):
		return "steel_mill"
	case strings.Contains(value, "仓库编码") || strings.Contains(value, "库位编码"):
		return "warehouse_code"
	case strings.Contains(value, "存放地") || strings.Contains(value, "仓库") || strings.Contains(value, "库房"):
		return "warehouse_name"
	case strings.Contains(value, "省"):
		return "province"
	case strings.Contains(value, "市"):
		return "city"
	case strings.Contains(value, "区") || strings.Contains(value, "县"):
		return "district"
	case strings.Contains(value, "地址"):
		return "address"
	case strings.Contains(value, "捆包") || strings.Contains(value, "包号") || strings.Contains(value, "bundle"):
		return "bundle_no"
	case strings.Contains(value, "起订") || strings.Contains(value, "最小"):
		return "min_order_qty"
	case strings.Contains(value, "库存") || strings.Contains(value, "重量") || strings.Contains(value, "数量") || strings.Contains(value, "吨数"):
		return "quantity"
	case strings.Contains(value, "单价") || strings.Contains(value, "价格") || strings.Contains(value, "报价"):
		return "unit_price"
	case strings.Contains(value, "备注") || strings.Contains(value, "说明") || strings.Contains(value, "描述"):
		return "description"
	default:
		return ""
	}
}

func importCell(row []string, headerMap map[string]int, key string) string {
	index, ok := headerMap[key]
	if !ok || index < 0 || index >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[index])
}

func isImportRowEmpty(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func inferProductImportLine(line string) (dto.ProductImportDraftItem, bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return dto.ProductImportDraftItem{}, false
	}
	spec := findSpecText(line)
	material := findMaterialText(line)
	quantity := parseImportNumberAfter(line, `(库存|重量|数量|吨数)[:：]?\s*`)
	unitPrice := parseImportNumberAfter(line, `(单价|价格|报价)[:：]?\s*`)
	steelMill := findSteelMillText(line)
	category := findCategoryText(line)

	if spec == "" && material == "" && category == "" {
		return dto.ProductImportDraftItem{}, false
	}

	product := dto.ProductImportProduct{
		GoodsType:     "spot",
		Category:      category,
		SubCategory:   category,
		SteelMill:     steelMill,
		Quantity:      quantity,
		UnitPrice:     unitPrice,
		MinOrderQty:   1,
		DeliveryDays:  3,
		QuoteValidity: 72,
		Spec: dto.ProductSpec{
			Material: material,
			Spec:     spec,
		},
		Description: line,
	}
	return dto.ProductImportDraftItem{Confidence: 0.45, Product: product}, true
}

func normalizeImportDraftItems(items []dto.ProductImportDraftItem) []dto.ProductImportDraftItem {
	result := make([]dto.ProductImportDraftItem, 0, len(items))
	for index, item := range items {
		item.RowNo = index + 1
		item.Product = normalizeImportProduct(item.Product)
		if item.Confidence <= 0 || item.Confidence > 1 {
			item.Confidence = 0.5
		}
		item.Warnings = compactImportStrings(append(item.Warnings, importProductWarnings(item.Product)...))
		result = append(result, item)
	}
	return result
}

func normalizeImportProduct(product dto.ProductImportProduct) dto.ProductImportProduct {
	product.Name = strings.TrimSpace(product.Name)
	product.GoodsType = normalizeImportGoodsType(product.GoodsType)
	product.Category = strings.TrimSpace(product.Category)
	product.SubCategory = strings.TrimSpace(product.SubCategory)
	product.SteelMill = strings.TrimSpace(product.SteelMill)
	product.BundleNo = strings.TrimSpace(product.BundleNo)
	product.Description = strings.TrimSpace(product.Description)
	product.Spec.Material = strings.TrimSpace(product.Spec.Material)
	product.Spec.Spec = strings.TrimSpace(product.Spec.Spec)
	product.Spec.Surface = strings.TrimSpace(product.Spec.Surface)
	product.Spec.Edge = strings.TrimSpace(product.Spec.Edge)
	product.Spec.Processing = strings.TrimSpace(product.Spec.Processing)
	product.WarehouseInfo.WarehouseName = strings.TrimSpace(product.WarehouseInfo.WarehouseName)
	product.WarehouseInfo.WarehouseCode = strings.TrimSpace(product.WarehouseInfo.WarehouseCode)
	product.WarehouseInfo.Province = strings.TrimSpace(product.WarehouseInfo.Province)
	product.WarehouseInfo.City = strings.TrimSpace(product.WarehouseInfo.City)
	product.WarehouseInfo.District = strings.TrimSpace(product.WarehouseInfo.District)
	product.WarehouseInfo.Address = strings.TrimSpace(product.WarehouseInfo.Address)

	if product.SubCategory == "" {
		product.SubCategory = product.Category
	}
	if product.Category == "" {
		product.Category = product.SubCategory
	}
	if product.Spec.Spec != "" {
		applySpecDimensions(&product)
	}
	if product.Name == "" {
		product.Name = composeImportProductName(product)
	}
	if product.MinOrderQty <= 0 && product.Quantity > 0 {
		product.MinOrderQty = 1
	}
	if product.DeliveryDays <= 0 {
		product.DeliveryDays = 3
	}
	if product.QuoteValidity <= 0 {
		product.QuoteValidity = 72
	}
	return product
}

func normalizeImportGoodsType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "期货", "futures", "future":
		return "futures"
	default:
		return "spot"
	}
}

func applySpecDimensions(product *dto.ProductImportProduct) {
	spec := strings.TrimSpace(product.Spec.Spec)
	normalized := strings.NewReplacer("×", "*", "x", "*", "X", "*", "Φ", "φ").Replace(spec)
	parts := strings.Split(normalized, "*")
	if len(parts) > 0 && product.Spec.HeightSpecValue <= 0 {
		product.Spec.HeightSpecValue = parseImportNumber(parts[0])
	}
	if len(parts) > 1 && product.Spec.WidthSpecValue <= 0 {
		product.Spec.WidthSpecValue = parseImportNumber(parts[1])
	}
	if len(parts) > 2 && product.Spec.LengthValue <= 0 {
		product.Spec.LengthValue = parseImportNumber(parts[2])
	}
	if product.Spec.HeightSpecType == "" {
		if strings.Contains(normalized, "φ") {
			product.Spec.HeightSpecType = "diameter"
		} else {
			product.Spec.HeightSpecType = "thickness"
		}
	}
	if product.Spec.WidthSpecType == "" {
		if strings.Contains(normalized, "φ") || strings.Contains(product.Category, "管") || strings.Contains(product.SubCategory, "管") {
			product.Spec.WidthSpecType = "wall_thickness"
		} else {
			product.Spec.WidthSpecType = "width"
		}
	}
	if product.Spec.LengthType == "" {
		switch {
		case strings.Contains(strings.ToUpper(spec), "C") || strings.Contains(spec, "卷"):
			product.Spec.LengthType = "coil"
		case product.Spec.LengthValue > 0:
			product.Spec.LengthType = "fixed"
		default:
			product.Spec.LengthType = "custom"
		}
	}
	if product.Spec.HeightSpecUnit == "" {
		product.Spec.HeightSpecUnit = "mm"
	}
	if product.Spec.WidthSpecUnit == "" {
		product.Spec.WidthSpecUnit = "mm"
	}
	if product.Spec.LengthUnit == "" {
		product.Spec.LengthUnit = "mm"
	}
	if product.Spec.Thickness == "" && product.Spec.HeightSpecType == "thickness" && product.Spec.HeightSpecValue > 0 {
		product.Spec.Thickness = formatSpecNumber(product.Spec.HeightSpecValue)
	}
	if product.Spec.Width == "" && product.Spec.WidthSpecType == "width" && product.Spec.WidthSpecValue > 0 {
		product.Spec.Width = formatSpecNumber(product.Spec.WidthSpecValue)
	}
	if product.Spec.Length == "" && product.Spec.LengthValue > 0 {
		product.Spec.Length = formatSpecNumber(product.Spec.LengthValue)
	}
}

func composeImportProductName(product dto.ProductImportProduct) string {
	parts := []string{product.Spec.Material, product.SubCategory, product.Spec.Spec}
	compact := compactImportStrings(parts)
	if len(compact) == 0 {
		return "待确认商品"
	}
	return strings.Join(compact, " ")
}

func importProductWarnings(product dto.ProductImportProduct) []string {
	warnings := make([]string, 0)
	if product.Category == "" {
		warnings = append(warnings, "缺少品种")
	}
	if product.Spec.Material == "" {
		warnings = append(warnings, "缺少牌号")
	}
	if product.Spec.Spec == "" {
		warnings = append(warnings, "缺少规格")
	}
	if product.SteelMill == "" {
		warnings = append(warnings, "缺少产地/钢厂")
	}
	if product.WarehouseInfo.WarehouseName == "" {
		warnings = append(warnings, "缺少存放地")
	}
	if product.Quantity <= 0 {
		warnings = append(warnings, "缺少库存数量")
	}
	if product.UnitPrice <= 0 {
		warnings = append(warnings, "缺少单价")
	}
	return warnings
}

func parseImportNumber(value string) float64 {
	value = strings.ReplaceAll(strings.TrimSpace(value), ",", "")
	if value == "" {
		return 0
	}
	matches := regexp.MustCompile(`[-+]?\d+(?:\.\d+)?`).FindString(value)
	if matches == "" {
		return 0
	}
	parsed, _ := strconv.ParseFloat(matches, 64)
	return parsed
}

func parseImportNumberAfter(line, prefixPattern string) float64 {
	re := regexp.MustCompile(prefixPattern + `([-+]?\d+(?:\.\d+)?)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 3 {
		return 0
	}
	return parseImportNumber(matches[2])
}

func findSpecText(line string) string {
	re := regexp.MustCompile(`(?i)(?:φ)?\d+(?:\.\d+)?\s*[×xX*]\s*\d+(?:\.\d+)?(?:\s*[×xX*]\s*(?:\d+(?:\.\d+)?|C))?`)
	return strings.TrimSpace(re.FindString(line))
}

func findMaterialText(line string) string {
	patterns := []string{
		`Q\d{3}[A-Z]?`,
		`[A-Z]{1,5}\d{2,4}[A-Z0-9]*`,
	}
	for _, pattern := range patterns {
		match := regexp.MustCompile(pattern).FindString(strings.ToUpper(line))
		if match != "" {
			return match
		}
	}
	return ""
}

func findSteelMillText(line string) string {
	re := regexp.MustCompile(`(宝钢|首钢|鞍钢|本钢|唐钢|邯钢|马钢|武钢|柳钢|包钢|日照|山钢|梅钢|太钢|攀钢|酒钢|承钢|宁钢|新余)`)
	return strings.TrimSpace(re.FindString(line))
}

func findCategoryText(line string) string {
	candidates := []string{"热轧板卷", "冷轧板卷", "中厚板", "镀锌卷", "热镀锌", "酸洗板卷", "彩涂板卷", "电工钢", "圆钢", "线材", "钢管", "板卷"}
	for _, candidate := range candidates {
		if strings.Contains(line, candidate) {
			return candidate
		}
	}
	return ""
}

func compactImportStrings(values []string) []string {
	result := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

type productImportModelResponse struct {
	Items    []productImportModelItem `json:"items"`
	Warnings []string                 `json:"warnings"`
}

type productImportModelItem struct {
	Name          string   `json:"name"`
	GoodsType     string   `json:"goods_type"`
	Category      string   `json:"category"`
	SubCategory   string   `json:"sub_category"`
	Material      string   `json:"material"`
	Spec          string   `json:"spec"`
	SteelMill     string   `json:"steel_mill"`
	WarehouseName string   `json:"warehouse_name"`
	WarehouseCode string   `json:"warehouse_code"`
	Province      string   `json:"province"`
	City          string   `json:"city"`
	District      string   `json:"district"`
	Address       string   `json:"address"`
	BundleNo      string   `json:"bundle_no"`
	Quantity      float64  `json:"quantity"`
	UnitPrice     float64  `json:"unit_price"`
	MinOrderQty   float64  `json:"min_order_qty"`
	Description   string   `json:"description"`
	Confidence    float64  `json:"confidence"`
	Warnings      []string `json:"warnings"`
}

func parseProductImportWithModel(ctx context.Context, source productImportSource, hint string) ([]dto.ProductImportDraftItem, []string, bool, error) {
	modelCfg := config.GetAIProcurement()
	if !modelCfg.ModelParsingEnabled {
		return nil, nil, false, nil
	}
	apiKey := modelCfg.APIKey()
	if apiKey == "" || modelCfg.BaseURL == "" || modelCfg.Model == "" {
		return nil, []string{"AI 模型配置不完整，已使用规则解析"}, false, nil
	}

	prompt := productImportModelPrompt(source, hint)
	messageContent := any(prompt)
	if source.IsImage {
		messageContent = []productImportMessagePart{
			{Type: "text", Text: prompt},
			{Type: "image_url", ImageURL: &productImportImageURL{URL: source.DataURL}},
		}
	}
	payload := productImportChatCompletionRequest{
		Model: modelCfg.Model,
		Messages: []productImportChatMessage{
			{Role: "system", Content: productImportSystemPrompt()},
			{Role: "user", Content: messageContent},
		},
		Temperature:         modelCfg.Temperature,
		TopP:                modelCfg.TopP,
		MaxCompletionTokens: modelCfg.MaxTokens,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, true, err
	}

	timeout := time.Duration(modelCfg.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	requestCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, productImportModelURL(modelCfg.BaseURL), bytes.NewReader(body))
	if err != nil {
		return nil, nil, true, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, true, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, nil, true, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, nil, true, fmt.Errorf("AI 商品导入解析失败: status %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	content, err := extractProductImportModelContent(responseBody)
	if err != nil {
		return nil, nil, true, err
	}
	content = normalizeProductImportJSONContent(content)
	var parsed productImportModelResponse
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return nil, nil, true, err
	}

	items := make([]dto.ProductImportDraftItem, 0, len(parsed.Items))
	for _, item := range parsed.Items {
		items = append(items, dto.ProductImportDraftItem{
			Confidence: item.Confidence,
			Warnings:   item.Warnings,
			Product: dto.ProductImportProduct{
				Name:        item.Name,
				GoodsType:   item.GoodsType,
				Category:    item.Category,
				SubCategory: item.SubCategory,
				SteelMill:   item.SteelMill,
				BundleNo:    item.BundleNo,
				Quantity:    item.Quantity,
				UnitPrice:   item.UnitPrice,
				MinOrderQty: item.MinOrderQty,
				Description: item.Description,
				Spec: dto.ProductSpec{
					Material: item.Material,
					Spec:     item.Spec,
				},
				WarehouseInfo: dto.WarehouseInfo{
					WarehouseName: item.WarehouseName,
					WarehouseCode: item.WarehouseCode,
					Province:      item.Province,
					City:          item.City,
					District:      item.District,
					Address:       item.Address,
				},
				DeliveryDays:  3,
				QuoteValidity: 72,
			},
		})
	}
	return normalizeImportDraftItems(items), parsed.Warnings, true, nil
}

type productImportChatCompletionRequest struct {
	Model               string                     `json:"model"`
	Messages            []productImportChatMessage `json:"messages"`
	Temperature         float64                    `json:"temperature,omitempty"`
	TopP                float64                    `json:"top_p,omitempty"`
	MaxCompletionTokens int                        `json:"max_completion_tokens,omitempty"`
}

type productImportChatMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type productImportMessagePart struct {
	Type     string                 `json:"type"`
	Text     string                 `json:"text,omitempty"`
	ImageURL *productImportImageURL `json:"image_url,omitempty"`
}

type productImportImageURL struct {
	URL string `json:"url"`
}

func productImportSystemPrompt() string {
	return strings.Join([]string{
		"你是钢材B2B平台的卖家商品台账解析器。",
		"只输出JSON，不要输出解释、Markdown或代码块。",
		"JSON必须是对象，包含 items 和 warnings。",
		"items 是数组，每个元素包含 name, goods_type, category, sub_category, material, spec, steel_mill, warehouse_name, warehouse_code, province, city, district, address, bundle_no, quantity, unit_price, min_order_qty, description, confidence, warnings。",
		"字段缺失时填空字符串或0；confidence 是0到1数字；warnings 是字符串数组。",
		"category/品种、material/牌号、spec/规格、steel_mill/产地、warehouse_name/存放地必须尽量从台账或截图中抽取。",
	}, "\n")
}

func productImportModelPrompt(source productImportSource, hint string) string {
	lines := []string{
		"请解析卖家上传的商品台账或截图，提取待批量添加的钢材商品。",
		"规格如 3.0*1500*C、3*1500*6000、Φ25*6*12000 要完整保留到 spec。",
		"数量默认单位为吨，单价默认单位为元/吨。",
	}
	if hint = strings.TrimSpace(hint); hint != "" {
		lines = append(lines, "商家补充说明: "+hint)
	}
	if !source.IsImage {
		lines = append(lines, "台账文本如下:", source.Text)
	}
	return strings.Join(lines, "\n")
}

func productImportModelURL(baseURL string) string {
	baseURL = strings.TrimSpace(baseURL)
	if strings.Contains(baseURL, "/chat/completions") {
		return baseURL
	}
	return strings.TrimRight(baseURL, "/") + "/chat/completions"
}

func extractProductImportModelContent(data []byte) (string, error) {
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("AI 商品导入解析未返回结果")
	}
	content := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if content == "" {
		return "", fmt.Errorf("AI 商品导入解析结果为空")
	}
	return content, nil
}

func normalizeProductImportJSONContent(content string) string {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start >= 0 && end > start {
		return content[start : end+1]
	}
	return content
}
