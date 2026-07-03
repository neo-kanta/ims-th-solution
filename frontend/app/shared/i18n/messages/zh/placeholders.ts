export const zhPlaceholderMessages = {
  placeholders: {
    approvalWorkflow: {
      title: "审批工作流",
      description: "审批工作流功能即将推出。",
    },
    approvalConfiguration: {
      title: "审批配置",
      description: "审批配置功能即将推出。",
    },
    analysisReports: {
      title: "分析报告",
      description: "分析报告功能即将推出。",
    },
    investmentDecisions: {
      title: "投资决策",
      description: "投资决策功能即将推出。",
    },
    executionOrders: {
      title: "交易執行與委託",
      description: "登錄交易並檢視執行狀態。",
      notConfiguredTitle: "交易執行以基金為單位處理",
      notConfiguredCopy:
        "後端尚未提供全域之委託管理 API。如需登錄真實的 BUY / SELL 交易,請於目標基金的「Operation」分頁進行;後端會在過帳時再次執行 Pre-trade 合規檢查,回傳成功即代表通過合規並寫入 Ledger。",
    },
    investmentReview: {
      title: "投资复核",
      description: "投资复核功能即将推出。",
    },
    leaveDelegation: {
      title: "请假与代理",
      description: "请假与代理功能即将推出。",
    },
    agentManagement: {
      title: "代理管理",
      description: "代理管理功能即将推出。",
    },
    accountManagement: {
      title: "账户管理",
      description: "账户管理功能即将推出。",
    },
    groupManagement: {
      title: "群组管理",
      description: "群组管理功能即将推出。",
    },
    accountLog: {
      title: "账户日志",
      description: "账户日志功能即将推出。",
    },
    auditLog: {
      title: "审计日志",
      description: "审计日志功能即将推出。",
    },
    notificationSettings: {
      title: "通知设置",
      description: "通知设置功能即将推出。",
    },
  },
} as const;
