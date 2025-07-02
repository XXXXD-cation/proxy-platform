package utils

import (
	"reflect"
	"testing"
	"time"
)

func TestStringUtils_BasicChecks(t *testing.T) {
	// 测试IsEmpty
	if !String.IsEmpty("") {
		t.Error("期望空字符串返回true")
	}
	if !String.IsEmpty("   ") {
		t.Error("期望只有空格的字符串返回true")
	}
	if String.IsEmpty("hello") {
		t.Error("期望非空字符串返回false")
	}

	// 测试IsNotEmpty
	if String.IsNotEmpty("") {
		t.Error("期望空字符串返回false")
	}
	if !String.IsNotEmpty("hello") {
		t.Error("期望非空字符串返回true")
	}

	// 测试Trim
	if String.Trim("  hello  ") != "hello" {
		t.Error("Trim函数未正确去除空格")
	}
}

func TestStringUtils_SearchOperations(t *testing.T) {
	// 测试Contains
	if !String.Contains("hello world", "world") {
		t.Error("Contains函数未正确检测子串")
	}
	if String.Contains("hello", "world") {
		t.Error("Contains函数错误检测不存在的子串")
	}

	// 测试ContainsIgnoreCase
	if !String.ContainsIgnoreCase("Hello World", "hello") {
		t.Error("ContainsIgnoreCase函数未正确忽略大小写")
	}

	// 测试StartsWith
	if !String.StartsWith("hello world", "hello") {
		t.Error("StartsWith函数未正确检测前缀")
	}

	// 测试EndsWith
	if !String.EndsWith("hello world", "world") {
		t.Error("EndsWith函数未正确检测后缀")
	}
}

func TestStringUtils_Transformations(t *testing.T) {
	// 测试Reverse
	if String.Reverse("hello") != "olleh" {
		t.Error("Reverse函数未正确反转字符串")
	}

	// 测试Truncate
	if String.Truncate("hello world", 5) != "hello..." {
		t.Error("Truncate函数未正确截断字符串")
	}
	if String.Truncate("hi", 5) != "hi" {
		t.Error("Truncate函数对短字符串处理错误")
	}

	// 测试PadLeft
	if String.PadLeft("123", 5, "0") != "00123" {
		t.Error("PadLeft函数未正确左填充")
	}

	// 测试PadRight
	if String.PadRight("123", 5, "0") != "12300" {
		t.Error("PadRight函数未正确右填充")
	}
}

func TestStringUtils_CaseConversions(t *testing.T) {
	// 测试CamelToSnake
	if String.CamelToSnake("HelloWorld") != "hello_world" {
		t.Error("CamelToSnake函数转换错误")
	}

	// 测试SnakeToCamel
	if String.SnakeToCamel("hello_world") != "helloWorld" {
		t.Error("SnakeToCamel函数转换错误")
	}
}

func TestNumberUtils_Validation(t *testing.T) {
	// 测试IsNumber
	if !Number.IsNumber("123.45") {
		t.Error("IsNumber函数未正确识别数字")
	}
	if Number.IsNumber("abc") {
		t.Error("IsNumber函数错误识别非数字")
	}

	// 测试IsInteger
	if !Number.IsInteger("123") {
		t.Error("IsInteger函数未正确识别整数")
	}
	if Number.IsInteger("123.45") {
		t.Error("IsInteger函数错误识别浮点数")
	}
}

func TestNumberUtils_Conversions(t *testing.T) {
	// 测试ToInt
	val, err := Number.ToInt("123")
	if err != nil || val != 123 {
		t.Error("ToInt函数转换错误")
	}

	// 测试ToInt64
	val64, err := Number.ToInt64("123456789")
	if err != nil || val64 != 123456789 {
		t.Error("ToInt64函数转换错误")
	}

	// 测试ToFloat64
	fval, err := Number.ToFloat64("123.45")
	if err != nil || fval != 123.45 {
		t.Error("ToFloat64函数转换错误")
	}
}

func TestNumberUtils_Operations(t *testing.T) {
	// 测试Round
	if Number.Round(3.14159, 2) != 3.14 {
		t.Error("Round函数四舍五入错误")
	}

	// 测试Max
	if Number.Max(5, 3) != 5 {
		t.Error("Max函数返回错误")
	}

	// 测试Min
	if Number.Min(5, 3) != 3 {
		t.Error("Min函数返回错误")
	}

	// 测试Abs
	if Number.Abs(-5) != 5 {
		t.Error("Abs函数返回错误")
	}
	if Number.Abs(5) != 5 {
		t.Error("Abs函数处理正数错误")
	}
}

func TestTimeUtils_BasicOperations(t *testing.T) {
	// 测试Now
	now := Time.Now()
	if now.IsZero() {
		t.Error("Now函数返回零值")
	}

	// 测试NowUnix
	unix := Time.NowUnix()
	if unix <= 0 {
		t.Error("NowUnix函数返回无效时间戳")
	}

	// 测试NowUnixMilli
	milli := Time.NowUnixMilli()
	if milli <= 0 {
		t.Error("NowUnixMilli函数返回无效毫秒时间戳")
	}
}

func TestTimeUtils_Formatting(t *testing.T) {
	testTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	// 测试Format
	if Time.Format(testTime, "2006-01-02") != "2023-01-01" {
		t.Error("Format函数格式化错误")
	}

	// 测试FormatDate
	if Time.FormatDate(testTime) != "2023-01-01" {
		t.Error("FormatDate函数格式化错误")
	}

	// 测试FormatDateTime
	if Time.FormatDateTime(testTime) != "2023-01-01 12:00:00" {
		t.Error("FormatDateTime函数格式化错误")
	}
}

func TestTimeUtils_Parsing(t *testing.T) {
	// 测试Parse
	parsed, err := Time.Parse("2006-01-02", "2023-01-01")
	if err != nil || parsed.Year() != 2023 {
		t.Error("Parse函数解析错误")
	}

	// 测试ParseDate
	date, err := Time.ParseDate("2023-01-01")
	if err != nil || date.Year() != 2023 {
		t.Error("ParseDate函数解析错误")
	}

	// 测试ParseDateTime
	datetime, err := Time.ParseDateTime("2023-01-01 12:00:00")
	if err != nil || datetime.Hour() != 12 {
		t.Error("ParseDateTime函数解析错误")
	}
}

func TestTimeUtils_Calculations(t *testing.T) {
	testTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	// 测试AddDays
	nextDay := Time.AddDays(testTime, 1)
	if nextDay.Day() != 2 {
		t.Error("AddDays函数计算错误")
	}

	// 测试AddHours
	nextHour := Time.AddHours(testTime, 1)
	if nextHour.Hour() != 13 {
		t.Error("AddHours函数计算错误")
	}

	// 测试AddMinutes
	nextMinute := Time.AddMinutes(testTime, 30)
	if nextMinute.Minute() != 30 {
		t.Error("AddMinutes函数计算错误")
	}

	// 测试DiffDays
	tomorrow := testTime.AddDate(0, 0, 1)
	if Time.DiffDays(tomorrow, testTime) != 1 {
		t.Error("DiffDays函数计算错误")
	}
}

func TestCryptoUtils(t *testing.T) {
	// 测试SHA256
	sha256Hash := Crypto.SHA256("hello")
	expectedSHA256 := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if sha256Hash != expectedSHA256 {
		t.Errorf("SHA256函数计算错误，期望 %s，实际 %s", expectedSHA256, sha256Hash)
	}

	// 测试GenerateRandomString
	randomStr := Crypto.GenerateRandomString(10)
	if len(randomStr) != 10 {
		t.Error("GenerateRandomString函数生成长度错误")
	}

	// 测试GenerateUUID
	uuid := Crypto.GenerateUUID()
	if uuid == "" {
		t.Error("GenerateUUID函数生成空UUID")
	}
}

func TestValidatorUtils(t *testing.T) {
	// 测试IsEmail
	if !Validator.IsEmail("test@example.com") {
		t.Error("IsEmail函数未正确验证邮箱")
	}
	if Validator.IsEmail("invalid-email") {
		t.Error("IsEmail函数错误验证无效邮箱")
	}

	// 测试IsPhone
	if !Validator.IsPhone("13800138000") {
		t.Error("IsPhone函数未正确验证手机号")
	}
	if Validator.IsPhone("12345") {
		t.Error("IsPhone函数错误验证无效手机号")
	}

	// 测试IsURL
	if !Validator.IsURL("https://example.com") {
		t.Error("IsURL函数未正确验证URL")
	}
	if Validator.IsURL("invalid-url") {
		t.Error("IsURL函数错误验证无效URL")
	}

	// 测试IsIP
	if !Validator.IsIP("192.168.1.1") {
		t.Error("IsIP函数未正确验证IP地址")
	}
	if Validator.IsIP("256.256.256.256") {
		t.Error("IsIP函数错误验证无效IP地址")
	}
}

func TestJSONUtils(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	testData := TestStruct{Name: "test", Age: 25}

	// 测试ToJSON
	jsonStr, err := JSON.ToJSON(testData)
	if err != nil {
		t.Errorf("ToJSON函数转换失败: %v", err)
	}
	if jsonStr == "" {
		t.Error("ToJSON函数返回空字符串")
	}

	// 测试FromJSON
	var result TestStruct
	err = JSON.FromJSON(jsonStr, &result)
	if err != nil {
		t.Errorf("FromJSON函数解析失败: %v", err)
	}
	if result.Name != "test" || result.Age != 25 {
		t.Error("FromJSON函数解析结果错误")
	}

	// 测试ToJSONPretty
	prettyJSON, err := JSON.ToJSONPretty(testData)
	if err != nil {
		t.Errorf("ToJSONPretty函数转换失败: %v", err)
	}
	if len(prettyJSON) <= len(jsonStr) {
		t.Error("ToJSONPretty函数未正确格式化")
	}

	// 测试IsValidJSON
	if !JSON.IsValidJSON(jsonStr) {
		t.Error("IsValidJSON函数验证有效JSON失败")
	}
	if JSON.IsValidJSON("invalid json") {
		t.Error("IsValidJSON函数错误验证无效JSON")
	}
}

func TestSliceUtils(t *testing.T) {
	// 测试Contains
	slice := []string{"a", "b", "c"}
	if !Slice.Contains(slice, "b") {
		t.Error("Contains函数未正确检测元素")
	}
	if Slice.Contains(slice, "d") {
		t.Error("Contains函数错误检测不存在元素")
	}

	// 测试ContainsInt
	intSlice := []int{1, 2, 3}
	if !Slice.ContainsInt(intSlice, 2) {
		t.Error("ContainsInt函数未正确检测元素")
	}
	if Slice.ContainsInt(intSlice, 4) {
		t.Error("ContainsInt函数错误检测不存在元素")
	}

	// 测试Remove
	result := Slice.Remove(slice, "b")
	expected := []string{"a", "c"}
	if !reflect.DeepEqual(result, expected) {
		t.Error("Remove函数未正确移除元素")
	}

	// 测试RemoveInt
	intResult := Slice.RemoveInt(intSlice, 2)
	expectedInt := []int{1, 3}
	if !reflect.DeepEqual(intResult, expectedInt) {
		t.Error("RemoveInt函数未正确移除元素")
	}

	// 测试Unique
	duplicateSlice := []string{"a", "b", "a", "c", "b"}
	uniqueResult := Slice.Unique(duplicateSlice)
	expectedUnique := []string{"a", "b", "c"}
	if !reflect.DeepEqual(uniqueResult, expectedUnique) {
		t.Error("Unique函数未正确去重")
	}

	// 测试UniqueInt
	duplicateIntSlice := []int{1, 2, 1, 3, 2}
	uniqueIntResult := Slice.UniqueInt(duplicateIntSlice)
	expectedUniqueInt := []int{1, 2, 3}
	if !reflect.DeepEqual(uniqueIntResult, expectedUniqueInt) {
		t.Error("UniqueInt函数未正确去重")
	}
}

func TestTernaryFunctions(t *testing.T) {
	// 测试Ternary
	result := Ternary(true, "yes", "no")
	if result != "yes" {
		t.Error("Ternary函数true条件返回错误")
	}

	result = Ternary(false, "yes", "no")
	if result != "no" {
		t.Error("Ternary函数false条件返回错误")
	}

	// 测试TernaryString
	strResult := TernaryString(true, "yes", "no")
	if strResult != "yes" {
		t.Error("TernaryString函数true条件返回错误")
	}

	// 测试TernaryInt
	intResult := TernaryInt(true, 1, 0)
	if intResult != 1 {
		t.Error("TernaryInt函数true条件返回错误")
	}
}

func TestRetry(t *testing.T) {
	// 测试成功情况
	attempts := 0
	err := Retry(3, time.Millisecond, func() error {
		attempts++
		if attempts < 2 {
			return nil
		}
		return nil
	})

	if err != nil {
		t.Errorf("Retry函数在成功情况下返回错误: %v", err)
	}

	// 测试失败情况
	failAttempts := 0
	err = Retry(3, time.Millisecond, func() error {
		failAttempts++
		return &testError{"test error"}
	})

	if err == nil {
		t.Error("Retry函数在失败情况下应该返回错误")
	}

	if failAttempts != 3 {
		t.Errorf("期望重试3次，实际重试%d次", failAttempts)
	}
}

// 辅助测试错误类型
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
