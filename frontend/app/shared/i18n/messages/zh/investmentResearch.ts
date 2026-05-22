import type { enInvestmentResearchMessages } from "./investmentResearch";

export const zhInvestmentResearchMessages = {
  investmentResearch: {
    // Pages / Headers
    title: "投资研究",
    newReport: "新建投资研究报告",
    editReport: "编辑投资研究报告",
    researchReport: "投资研究报告",
    descriptionList: "分析师研究报告，包含建议和生命周期跟踪。",
    descriptionNew: "在草稿报告中记录分析师的建议。",
    loading: "加载投资研究报告中…",
    loadingList: "加载投资研究报告中…",
    untitledReport: "未命名报告",
    identification: "识别信息",
    reportContent: "报告内容",

    // Labels & Table columns / Form Fields
    reportNo: "报告编号",
    reportDate: "报告日期",
    reportDateFrom: "报告日期起",
    reportDateTo: "报告日期止",
    effectiveDate: "生效日期",
    instrument: "工具",
    instrumentCode: "工具代码",
    instrumentName: "工具名称",
    instrumentType: "工具类型",
    market: "市场",
    currency: "货币",
    recommendation: "建议",
    reportStatus: "报告状态",
    review: "评审",
    reviewStatus: "评审状态",
    titleField: "标题",
    ownerUserId: "所有者用户ID",
    authorUserId: "作者用户ID",
    applicableContractId: "适用合同ID",
    applicableContract: "适用合同",
    ownerAuthor: "所有者 / 作者",
    typeMarketCurrency: "类型 / 市场 / 货币",
    companyOverview: "公司概览",
    companyOutlook: "公司展望",
    financialStatus: "财务状况",
    esgComment: "ESG评级",
    investmentAnalysis: "投资分析",
    lifecycleNotes: "生命周期备注",
    rejectionReason: "拒绝原因",
    postSubmissionNote: "提交后备注",
    created: "创建时间",
    lastUpdated: "最后更新",
    timezone: "时区",
    timezoneValue: "亚洲/曼谷 (UTC+7)",
    by: "由",
    keepCurrent: "保持当前",

    // Filters
    searchPlaceholder: "报告编号、代码、标题",
    search: "搜索",
    all: "全部",
    selectPlaceholder: "请选择…",

    // Form help text and errors
    fieldLimits: {
      maxLength: "最多 {max} 个字符",
      minLength: "投资分析必须至少为 {min} 个字符",
      uuid: "必须是 UUID",
      uuidOptional: "必须是 UUID（或留空）",
      currencyCode: "必须是 3 个大写字母（例如 THB）",
    },
    required: "必填",
    characterCounter: "{count} / {max}",
    characterCounterAnalysis: "{count} / {min} 个字符",
    minimumAnalysisPlaceholder: "最少 {min} 个字符",
    uuidOptionalPlaceholder: "UUID (可选)",

    // Actions & Buttons
    actions: {
      apply: "套用",
      reset: "重置",
      back: "返回",
      edit: "编辑",
      submit: "提交",
      cancelSubmission: "取消提交",
      delete: "删除",
      cancel: "取消",
      createReport: "创建报告",
      saveChanges: "保存修改",
      prev: "上一页",
      next: "下一页",
      shortcutHintCreate: "按 ⌘/Ctrl + Enter 创建",
      shortcutHintSave: "按 ⌘/Ctrl + Enter 保存",
    },

    // Confirm dialogs
    confirm: {
      delete: {
        title: "删除此投资研究报告？",
        description: "软删除该报告会将其从分析师列表和评审队列中移除。审计追踪仍将保留。该操作无法在界面上撤销。",
        confirmLabel: "删除报告",
      },
      submit: {
        title: "提交报告进行评审？",
        description: "报告将从“未提交”状态转移到“已提交”状态，并对评审组可见。您可以在评审完成前取消提交。",
        confirmLabel: "提交",
      },
      cancelSubmit: {
        title: "取消提交？",
        description: "报告将回到“未提交”状态，并从评审队列中消失。评审员已有的意见将保留。",
        confirmLabel: "取消提交",
      },
    },

    // Statuses
    status: {
      draft: "草稿",
      active: "有效",
      expired: "过期",
      rejected: "拒绝",
    },
    reviewStatusValues: {
      notSubmitted: "未提交",
      submitted: "已提交",
      reviewCompleted: "评审完成",
    },
    recommendationValues: {
      buy: "买入",
      sell: "卖出",
      hold: "持有",
    },

    // Errors
    errors: {
      loadList: "加载投资研究报告失败",
      loadDetail: "加载投资研究报告失败",
      create: "创建投资研究报告失败",
      update: "更新投资研究报告失败",
      delete: "删除投资研究报告失败",
      submit: "提交投资研究报告失败",
      cancelSubmit: "取消提交投资研究报告失败",
    },

    // Empty state
    empty: {
      title: "暂无投资研究报告",
      description: "创建草稿以开始跟踪分析师建议。",
    },

    // Footer paging
    pager: {
      range: "第 {page} 页 · 显示 {count} / 共 {total}",
    },
  },
} as const satisfies typeof enInvestmentResearchMessages;
