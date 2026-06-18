export const zhApprovalMessages = {
  approval: {
    nav: {
      section: "审批",
      inbox: "审批收件箱",
      processes: "审批流程",
      groups: "审批组",
      teams: "审批团队",
    },
    inbox: {
      title: "审批收件箱",
      description: "分配给您的待审批事项。",
      count: "{total} 项待处理任务",
    },
    request: {
      title: "审批请求",
      description: "查看请求、时间线并签核。",
      loading: "正在加载请求…",
    },
    actions: {
      refresh: "刷新",
      backToInbox: "返回收件箱",
    },
    errors: {
      noPermission: "您没有执行此操作的权限。",
      notFound: "未找到审批请求。",
    },
    config: {
      title: "审批配置",
      description: "配置审批流程、组和团队。",
      processes: {
        title: "审批流程",
        description: "按类型和合同配置审批流程和阶段。",
        new: "新建流程",
        listTitle: "流程",
        empty: "尚未配置审批流程。",
      },
      groups: {
        title: "审批组",
        description: "管理可复用的审批组和成员。",
        new: "新建组",
        edit: "编辑组",
        listTitle: "组",
        empty: "尚未配置审批组。",
        membersTitle: "成员 — ",
      },
      teams: {
        title: "审批团队",
        description: "按合同/基金配置审批团队。",
        new: "新建团队",
        listTitle: "团队",
        empty: "尚未配置审批团队。",
      },
    },
  },
} as const;
