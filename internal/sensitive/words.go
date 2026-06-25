package sensitive

// Category 敏感词分类
const (
	CategoryPolitical = "political" // 政治敏感
	CategoryViolence  = "violence"  // 暴力
	CategoryPorn      = "porn"      // 色情
	CategorySelfHarm  = "self_harm" // 自伤
)

// Words 敏感词库：词 → 分类
// 共 ~50 个中文敏感词，覆盖政治 / 暴力 / 色情 / 自伤四类
var Words = map[string]string{
	// ── 政治敏感类 (political) ──────────────────────
	"颠覆国家政权": CategoryPolitical,
	"分裂国家":   CategoryPolitical,
	"煽动民族仇恨": CategoryPolitical,
	"邪教组织":   CategoryPolitical,
	"非法集会":   CategoryPolitical,
	"反党":     CategoryPolitical,
	"反共":     CategoryPolitical,
	"恐怖组织":   CategoryPolitical,
	"煽动颠覆":   CategoryPolitical,

	// ── 暴力类 (violence) ───────────────────────────
	"杀了你":    CategoryViolence,
	"弄死你":    CategoryViolence,
	"砍死":     CategoryViolence,
	"碎尸":     CategoryViolence,
	"炸了":     CategoryViolence,
	"灭了你":    CategoryViolence,
	"打死你":    CategoryViolence,
	"血腥":     CategoryViolence,
	"屠杀":     CategoryViolence,
	"行刑":     CategoryViolence,
	"虐杀":     CategoryViolence,
	"肢解":     CategoryViolence,
	"绑架":     CategoryViolence,
	"投毒":     CategoryViolence,
	"纵火":     CategoryViolence,

	// ── 色情类 (porn) ──────────────────────────────
	"裸聊":     CategoryPorn,
	"约炮":     CategoryPorn,
	"卖淫":     CategoryPorn,
	"嫖娼":     CategoryPorn,
	"性交":     CategoryPorn,
	"淫秽":     CategoryPorn,
	"色情服务":   CategoryPorn,
	"招嫖":     CategoryPorn,
	"一夜情":    CategoryPorn,
	"黄片":     CategoryPorn,
	"成人视频":   CategoryPorn,
	"包养":     CategoryPorn,
	"上门服务":   CategoryPorn,
	"三陪":     CategoryPorn,

	// ── 自伤类 (self_harm) ─────────────────────────
	"想死":     CategorySelfHarm,
	"不想活了":   CategorySelfHarm,
	"活不下去":   CategorySelfHarm,
	"自残":     CategorySelfHarm,
	"割腕":     CategorySelfHarm,
	"自杀":     CategorySelfHarm,
	"去死":     CategorySelfHarm,
	"结束生命":   CategorySelfHarm,
	"了断":     CategorySelfHarm,
	"跳楼":     CategorySelfHarm,
	"上吊":     CategorySelfHarm,
	"吞药":     CategorySelfHarm,
	"安眠药":    CategorySelfHarm,
	"没有意义活着": CategorySelfHarm,
	"生无可恋":   CategorySelfHarm,
	"死了一了百了": CategorySelfHarm,
}

// SelfHarmHelpline 自伤干预提示语
const SelfHarmHelpline = "你的文字透露出你可能正在经历艰难时刻。请记住，你并不孤单。" +
	"全国24小时免费心理危机干预热线：010-82951332。" +
	"希望24热线（北京）：400-161-9995。"
