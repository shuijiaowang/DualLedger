package model

// Category 系统级分类（MVP 静态常量；V2 再支持用户自定义）
type Category struct {
	Code       string `json:"code"`
	ParentCode string `json:"parent_code,omitempty"`
	Name       string `json:"name"`
	Kind       string `json:"kind"` // INCOME / EXPENSE / TRANSFER / OTHER
	Icon       string `json:"icon,omitempty"`
	Sort       int    `json:"sort"`
	Source     string `json:"source"` // system / user（MVP 恒为 system）
}

// PresetCategories MVP 预设分类清单，对齐数据库设计.md §5.2
// 后续放开自定义时，把 source=user 的加进来一起展示即可。
var PresetCategories = []Category{
	// ---- EXPENSE ----
	{Code: "food", Name: "餐饮", Kind: CategoryKindExpense, Sort: 10, Source: "system"},
	{Code: "food.breakfast", ParentCode: "food", Name: "早餐", Kind: CategoryKindExpense, Sort: 11, Source: "system"},
	{Code: "food.lunch", ParentCode: "food", Name: "午餐", Kind: CategoryKindExpense, Sort: 12, Source: "system"},
	{Code: "food.dinner", ParentCode: "food", Name: "晚餐", Kind: CategoryKindExpense, Sort: 13, Source: "system"},
	{Code: "food.takeout", ParentCode: "food", Name: "外卖", Kind: CategoryKindExpense, Sort: 14, Source: "system"},
	{Code: "food.snack", ParentCode: "food", Name: "零食饮料", Kind: CategoryKindExpense, Sort: 15, Source: "system"},

	{Code: "transport", Name: "交通", Kind: CategoryKindExpense, Sort: 20, Source: "system"},
	{Code: "transport.public", ParentCode: "transport", Name: "公共交通", Kind: CategoryKindExpense, Sort: 21, Source: "system"},
	{Code: "transport.taxi", ParentCode: "transport", Name: "打车", Kind: CategoryKindExpense, Sort: 22, Source: "system"},
	{Code: "transport.fuel", ParentCode: "transport", Name: "加油", Kind: CategoryKindExpense, Sort: 23, Source: "system"},

	{Code: "shopping", Name: "购物", Kind: CategoryKindExpense, Sort: 30, Source: "system"},
	{Code: "shopping.daily", ParentCode: "shopping", Name: "日用品", Kind: CategoryKindExpense, Sort: 31, Source: "system"},
	{Code: "shopping.clothes", ParentCode: "shopping", Name: "服饰", Kind: CategoryKindExpense, Sort: 32, Source: "system"},
	{Code: "shopping.digital", ParentCode: "shopping", Name: "电子产品", Kind: CategoryKindExpense, Sort: 33, Source: "system"},

	{Code: "home", Name: "居家", Kind: CategoryKindExpense, Sort: 40, Source: "system"},
	{Code: "home.rent_utility", ParentCode: "home", Name: "房租水电", Kind: CategoryKindExpense, Sort: 41, Source: "system"},
	{Code: "home.telecom", ParentCode: "home", Name: "通讯网络", Kind: CategoryKindExpense, Sort: 42, Source: "system"},

	{Code: "entertainment", Name: "娱乐", Kind: CategoryKindExpense, Sort: 50, Source: "system"},
	{Code: "entertainment.subscription", ParentCode: "entertainment", Name: "订阅会员", Kind: CategoryKindExpense, Sort: 51, Source: "system"},
	{Code: "entertainment.game", ParentCode: "entertainment", Name: "游戏", Kind: CategoryKindExpense, Sort: 52, Source: "system"},
	{Code: "entertainment.travel", ParentCode: "entertainment", Name: "出行", Kind: CategoryKindExpense, Sort: 53, Source: "system"},

	{Code: "health", Name: "医疗健康", Kind: CategoryKindExpense, Sort: 60, Source: "system"},
	{Code: "misc.expense", Name: "其他支出", Kind: CategoryKindExpense, Sort: 99, Source: "system"},

	// ---- INCOME ----
	{Code: "income.salary", Name: "工资", Kind: CategoryKindIncome, Sort: 110, Source: "system"},
	{Code: "income.bonus", Name: "奖金", Kind: CategoryKindIncome, Sort: 120, Source: "system"},
	{Code: "income.finance", Name: "理财", Kind: CategoryKindIncome, Sort: 130, Source: "system"},
	{Code: "income.gift", Name: "红包礼金", Kind: CategoryKindIncome, Sort: 140, Source: "system"},
	{Code: "misc.income", Name: "其他收入", Kind: CategoryKindIncome, Sort: 199, Source: "system"},

	// ---- 内部流转 ----
	{Code: "internal", Name: "内部流转", Kind: CategoryKindOther, Sort: 900, Source: "system"},
}

// FindCategory 根据 code 找分类（nil 表示不存在）
func FindCategory(code string) *Category {
	for i := range PresetCategories {
		if PresetCategories[i].Code == code {
			return &PresetCategories[i]
		}
	}
	return nil
}

// CategoryExists 判定 code 是否合法（包含空串 = 允许为空，留给前端判空）
func CategoryExists(code string) bool {
	if code == "" {
		return true
	}
	return FindCategory(code) != nil
}
