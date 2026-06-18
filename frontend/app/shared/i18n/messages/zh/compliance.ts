export const zhComplianceMessages = {
  compliance: {
    // Navigation
    nav: {
      section: "合規管理",
      dashboard: "合規儀表板",
      rules: "規則庫",
      preTrade: "前置交易模擬器 (Pre-Trade)",
      postTrade: "後置交易違規處理 (Post-Trade)",
      exceptions: "前置交易豁免申請",
      audit: "審計追蹤",
      permissions: "權限矩陣",
      newRule: "新建規則",
    },

    // Shared
    common: {
      loading: "載入中…",
      retry: "重試",
      backendError: "後端回傳錯誤。",
      none: "—",
      notYetAvailable: "後端尚未支援此功能。",
    },

    // Dashboard
    dashboard: {
      title: "合規管理",
      description:
        "投資限制與準則總覽。前置交易模擬器能在不良委託單轉化為交易、結算或審計問題之前，及時予以攔截。",
      restrictedHint: "受限存取",
      headerActions: {
        recentChanges: "最近變更",
        openBreaches: "未處理違規",
        newRule: "新建規則",
      },
      tabs: {
        overview: "總覽",
        library: "規則庫",
        approvals: "審批",
        breaches: "違規",
        exceptions: "豁免",
        audit: "審計",
        settings: "設置",
      },
      kpi: {
        activeRules: "生效中規則",
        activeRulesSub: "已發佈且有效",
        draft: "草稿",
        draftSub: "尚未提交",
        draftHint:
          "目前後端不支援草稿（Draft）狀態；此處統計包含 effective_from 設在未來的未啟用規則。",
        pendingReview: "待審核",
        pendingReviewSub: "等待審批",
        pendingReviewHint:
          "目前後端不支援審批流程（Approval Flow）。",
        disabled: "已禁用",
        disabledSub: "近期停用",
        expired: "已過期",
        expiredSub: "已超過有效期限",
        highRisk: "高風險規則",
        highRiskSub: "嚴重程度 = BLOCK",
        totalRules: "規則總計",
        inactiveRules: "未啟用 / 排程中",
        recentBreaches: "最近違規",
        trendUnavailable:
          "無法提供週對週趨勢分析 — 目前缺乏歷史 KPI 記錄端點。",
      },
      recentChanges: {
        title: "最近規則變更",
        viewAll: "查看全部 →",
        empty: "目前無規則變更歷史紀錄。",
      },
      recentFailures: {
        title: "最近合規違規",
        description: "最新未處理或近期已解決的違規記錄。",
        empty: "目前無合規違規記錄。",
        inboxLink: "打開違規收件匣 →",
        reviewCta: "審查",
        preTrade: "前置交易",
        postTrade: "後置交易",
      },
      search: {
        placeholder: "搜尋規則：名稱、代碼、參數…",
        filterFund: "基金",
        filterAssetClass: "資產類別",
        filterRuleType: "規則類型",
        filterStatus: "狀態",
        filterSeverity: "嚴重程度",
        filterEffective: "生效時間 ▼",
        filterOwner: "擁有者",
        filterApproval: "審批狀態 ▼",
        backendNote:
          "後端僅支援篩選 rule_type_id + is_active。其餘Facet需待端點 GET /compliance/rules 擴展後支援。",
      },
      categories: {
        title: "類別",
        empty: "目前尚無分類規則摘要。",
      },
      highRiskRules: {
        title: "高風險規則",
        seeAll: "查看全部 →",
        empty: "無預設嚴重程度為 BLOCK 的規則。",
      },
      truncatedNotice:
        "顯示規則 1-{shown}（共 {total} 條）— 導出之統計數據僅基於此視窗內計算。",
      goToSimulator: "運行前置交易模擬器",
      goToRules: "打開規則庫",
    },

    // Empty states
    emptyState: {
      noRulesTitle: "未配置任何啟用的合規規則",
      noRulesSubtitle: "尚無可用於評估委託單的規則。",
      noRulesIntro:
        "合規檢查需至少配置一個啟用的規則實例。在設定規則前，前置交易模擬器無法提供有效的判斷結果。",
      setupChecklist: "設定檢查清單",
      step1:
        "確認已執行資料庫遷移 20260417000001_compliance__create_rules_tables (make migrate-up)。",
      step2:
        "請先透過 Swagger 格式的 API POST /compliance/rules 建立規則實例（規則生成介面將在第二階段上線）。",
      step3:
        "確認您的帳戶已具備 IRG_VIEW_RULES 與 WORKFLOW_EXECUTE 權限。",
      noRulesEvaluatedTitle: "前置交易檢查評估了 0 條規則",
      noRulesEvaluatedSubtitle:
        "後端已執行檢查，但此委託單不適用任何規則 — 此處若顯示通過（PASS）將會產生誤導。",
    },

    // Rule library
    rules: {
      title: "規則庫",
      description:
        "已配置的合規規則實例。第一階段為唯讀；編輯與生命週期操作將於第二階段推出。",
      addRuleUnavailable:
        "規則編輯器將於第二階段上線。目前請直接透過 API 建立規則。",
      table: {
        code: "規則代碼",
        name: "名稱",
        category: "類別",
        severity: "嚴重程度",
        status: "狀態",
        version: "版本",
        effective: "生效區間",
        owner: "擁有者",
        updated: "最後更新",
        empty: "未找到合規規則。",
      },
      filters: {
        title: "篩選器",
        ruleTypeId: "規則類型 ID",
        ruleTypeIdHint: "例如 concentration.single_issuer",
        isActive: "啟用狀態",
        any: "全部",
        activeOnly: "僅啟用",
        inactiveOnly: "僅未啟用",
        apply: "套用篩選",
        reset: "重設",
      },
    },

    // Pre-trade simulator
    preTrade: {
      title: "前置交易合規模擬器 (Pre-Trade)",
      description:
        "提交擬議之委託單；後端將在建立實際交易前執行所有適用的規則並回傳結果。若結果為 BLOCK，該委託將被禁止繼續執行。",
      readonlyTitle: "執行 Pre-trade 需要 WORKFLOW_EXECUTE 權限。",
      readonlyCopy:
        "您可以查看現行規則與既有結果,但目前角色無法執行新的模擬。請向管理員申請 WORKFLOW_EXECUTE 權限後再操作。",
      form: {
        sectionOrder: "擬議委託單",
        sectionContext: "投資組合內容",
        portfolio: "投資組合",
        portfolioPlaceholder: "選擇投資組合…",
        portfolioEmpty:
          "目前無您的帳戶可讀取的投資組合。請聯繫管理員授權存取範圍，或直接在下方貼入投資組合 ID (UUID)。",
        portfolioLoading: "載入投資組合中…",
        portfolioErrorPrefix: "無法載入投資組合：",
        portfolioIdFallback: "投資組合 ID (UUID — 備用)",
        portfolioId: "投資組合 ID",
        contract: "合約",
        contractPlaceholder: "選擇合約…",
        contractEmpty:
          "您的權限範圍內無合約。如果您知道 ID，請在下方貼入 UUID。",
        contractIdFallback: "合約 ID (UUID — 備用)",
        contractId: "合約 ID",
        businessDate: "交易日期",
        orderId: "委託單 ID",
        generate: "隨機生成",
        ticker: "證券代碼 (Ticker)",
        side: "交易方向",
        buy: "買入",
        sell: "賣出",
        quantity: "數量",
        price: "價格",
        currency: "幣別",
        currencyAutofill: "依選擇的投資組合自動填寫。",
        exchange: "交易所",
        submit: "運行前置交易檢查",
        running: "正在運行合規檢查…",
        invalid: "部分必填欄位缺失或格式不正確。",
        invalidUuid: "必須為 UUID 格式 (xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)。",
        invalidDecimal: "必須為正實數。",
        required: "必填。",
      },
      errors: {
        workflowExecuteRequired:
          "您僅能查看合規規則，不具備運行前置交易模擬檢查的 WORKFLOW_EXECUTE 權限。",
        unauthorizedRules:
          "登入逾時，請重新登入系統。",
      },
      result: {
        block: "已被攔截 (Blocked)",
        blockSummaryOne: "有 1 條規則阻擋了此委託單的提交。",
        blockSummaryMany: "有 {count} 條規則阻擋了此委託單的提交。",
        warn: "合規警告",
        warnSummaryOne: "有 1 條規則觸發了警告。",
        warnSummaryMany: "有 {count} 條規則觸發了警告。",
        pass: "所有檢查已通過",
        passSummary:
          "已評估 {count} 條規則，0 次違規。委託單符合規定。",
        durationLabel: "評估完成，耗時 {ms} 毫秒",
        groupIdLabel: "檢查群組",
        rulesEvaluated: "評估規則數",
        currentValue: "當前數值",
        afterOrder: "下單後預估",
        ruleLimit: "規則限制上限",
        difference: "差距",
        dataNotProvided:
          "後端未回傳數據 — 後端佐證資料物件為空。",
        evidenceLookupFailed:
          "無法從此檢查群組中取得違規佐證資料 ({error})。",
        explanationLabel: "未通過原因",
        suggestedCorrection: "建議修正方式",
        actions: "操作",
        recheck: "重新檢查",
        requestException: "申請豁免",
        requestExceptionDisabled:
          "目前後端尚不支援豁免申請流程。",
        viewRule: "查看規則詳情",
      },
    },

    // Phase 2 — Post-trade breach inbox
    postTrade: {
      title: "後置交易違規收件匣 (Post-Trade)",
      description:
        "記錄交易捕獲後或定期回放期間所產生的合規違規。支援手動跳過並填寫理由，但尚不支援完整的狀態流轉功能。",
      empty: "目前無違規紀錄。",
      filters: {
        title: "篩選器",
        status: "狀態",
        any: "全部",
        open: "未處理",
        overridden: "已跳过 (Overridden)",
        resolved: "已解決",
        ruleTypeId: "規則類型 ID",
        dateFrom: "起始日期",
        dateTo: "結束日期",
        portfolio: "投資組合",
        portfolioAny: "所有投資組合",
        portfolioId: "投資組合 ID (UUID 備用)",
        portfolioLoading: "載入投資組合中…",
        contract: "合約",
        contractAny: "所有合約",
        contractId: "合約 ID (UUID 備用)",
        apply: "套用",
        reset: "重設",
      },
      table: {
        rule: "投資規則",
        severity: "嚴重程度",
        verdict: "評估結果",
        status: "狀態",
        businessDate: "交易日期",
        portfolio: "投資組合",
        contract: "合約",
        message: "通知訊息",
        actions: "操作",
        override: "跳過違規…",
        viewGroup: "查看檢查群組",
        statusTransitionDisabled:
          "後端目前尚未支援狀態流轉操作。",
      },
      override: {
        title: "跳過違規事項",
        description:
          "跳過違規的記錄將被永久保存並納入審計。此操作僅限合規人員執行 (IRG_OVERRIDE_BREACH)。",
        reason: "跳過理由",
        reasonHelp:
          "長度至少為 10 個字元。此說明將保存在永久的跳過記錄中。",
        approvedBy: "協同審批人 (選填)",
        approvedByHelp:
          "如果您的控制架構需要雙人覆核 (Four-eyes principle)，請提供第二級審批人 UUID。",
        cancel: "取消",
        submit: "確認跳過",
        submitting: "正在保存數據…",
        success: "違規已跳過並記錄。",
        notOverridable:
          "該項目並非處於 OPEN 狀態。只有 OPEN 狀態的違規事項才允許執行跳過操作。",
      },
      groupDrawer: {
        title: "檢查群組",
        records: "已檢查項目 ({count})",
        breaches: "違規項目 ({count})",
        close: "關閉",
        loading: "載入檢查群組中…",
        empty: "後端回傳空資料封包。",
      },
    },

    // Phase 2 — Rule Detail
    detail: {
      backToLibrary: "返回規則庫",
      tabs: {
        overview: "總覽",
        logic: "評估邏輯",
        scope: "綁定範圍",
        test: "功能測試",
        approval: "審批歷史",
        audit: "變更歷史",
        settings: "設置",
      },
      header: {
        editUnavailable:
          "後端目前尚不支援編輯規則操作。",
        disableUnavailable:
          "後端目前尚不支援停用規則操作。",
      },
      rail: {
        status: "狀態",
        ruleTypeId: "規則類型 ID",
        owner: "所有權人",
        version: "使用中版本",
        category: "類別",
        defaultSeverity: "預設嚴重程度",
        timings: "執行檢查時點",
        scopes: "適用資產類別",
        overridable: "允許例外豁免",
        effectiveWindow: "生效區間",
        createdAt: "建立日期",
        updatedAt: "最後修改日期",
      },
      overview: {
        descriptionTitle: "描述",
        descriptionEmpty: "未提供任何詳細描述。",
        metadataDescription: "後端規則類型的工程描述說明",
      },
      logic: {
        title: "評估邏輯",
        description:
          "當前執行的規則參數。在此版本中，尚不支援直接在網頁上編輯生效版本。",
        parametersTitle: "規則參數",
        parametersUnavailable:
          "無法從當前的列表端點中獲取生效版本的參數 payload。",
      },
      scope: {
        title: "綁定範圍",
        scopesLabel: "此規則所限制綁定的資產類別或系統範圍",
        bindingsTitle: "規則綁定 (Bindings)",
        bindingsUnavailable:
          "目前後端尚無提供各綁定實例的覆寫詳情。",
      },
      test: {
        title: "測試此規則",
        description:
          "您可以透過前置交易模擬器對此規則進行測試。請選擇與此規則相同的投資組合/合約，並輸入測試交易資訊。",
        cta: "在前置交易模擬器中打開",
      },
      approval: {
        title: "審批歷史",
        unavailable:
          "後端目前不支援合規規則的審批流程（Approval Flow）。",
      },
      audit: {
        title: "變更歷史",
        unavailable:
          "目前尚未支援版本追蹤與合規變更審計記錄功能。",
      },
      settings: {
        title: "設置",
        unavailable:
          "目前尚未支援調整規則過期日或設定永久禁用規則。",
      },
    },

    // Phase 2 — Rule Builder
    builder: {
      title: "建立新合規規則",
      description:
        "執行 POST /compliance/rules 建立並啟用規則。目前後端尚未支援發起審批、同意/拒絕及手動停用等流程。",
      lifecycleNotice:
        "本精靈將會一步完成建立與啟用規則。草稿和審批的生命週期流程將於後端 API 健全後上線。",
      steps: {
        identity: "基本資訊",
        scope: "範圍配置",
        logic: "評估邏輯",
        message: "提示訊息",
        review: "檢查設定",
        submit: "提交保存",
      },
      identity: {
        name: "規則名稱",
        description: "規則描述",
        category: "類別",
        ruleTypeId: "規則類型 ID",
        ruleTypeIdHelp:
          "直接連結後端的規則標識碼。請從列表中選取，或手動輸入。",
        catalogUnavailable:
          "無法從後端載入規則類型列表。下方的選項僅為對模擬後端支援規則類型之鏡像。",
      },
      scope: {
        effectiveFrom: "生效日期",
        effectiveTo: "到期日期 (選填)",
        isActive: "立即啟用此規則",
        isActiveHelp:
          "若取消核取，規則建立後將處於停用狀態且不會參與下單檢核。",
      },
      logic: {
        parametersTitle: "評估參數 (JSON)",
        parametersHelp:
          "符合此規則類型 ParameterSchema 的 JSON 參數內容。下方展示預設的參數範例格式。",
        invalidJson: "參數內容必須為合法的 JSON 格式。",
      },
      message: {
        title: "提示訊息說明",
        notice:
          "合規違規訊息將由後端引擎自動產生，目前尚不支援在實例層面覆寫專屬訊息。",
        changeReason: "變更理由說明",
        changeReasonHelp:
          "將記錄在版本歷史中，便於回溯說明此規則建立的原因。",
      },
      review: {
        title: "檢查規則配置",
        summary: "規則摘要資訊",
        payloadTitle: "Request Payload 內容",
      },
      submit: {
        title: "確認並提交保存",
        notice:
          "確認無誤後將發送 POST /compliance/rules 請求，新規則將即刻生效並出現在規則庫列表中。",
        cta: "建立此規則",
        submitting: "正在建立…",
        success: "已成功建立合規規則並保存。",
        viewRule: "開啟規則",
        backToLibrary: "返回規則庫列表",
      },
      lifecycleDisabled: {
        submitForApproval: "提交審批 (系統暫不支援)",
        approve: "批准啟用 (系統暫不支援)",
        reject: "拒絕批准 (系統暫不支援)",
        disable: "暫停使用 (系統暫不支援)",
        archive: "封存記錄 (系統暫不支援)",
      },
      validation: {
        required: "此欄位為必填項目。",
        nameTooShort: "長度至少需 3 個字元。",
        invalidDate: "日期格式不正確，請使用 YYYY-MM-DD 格式。",
      },
    },

    // Phase 2 — Exception Flow
    exceptions: {
      title: "前置交易例外豁免申請 (Pre-Trade Exceptions)",
      description:
        "在委託單下單前，針對被阻擋 (BLOCK) 的結果向合規部門發起暫時豁免申請，有別於後置交易的違規跳過。",
      notice:
        "後端目前尚未支援豁免相關端點。下方的申請表單僅為 UI 展示預覽，提交按鈕已設為禁用狀態。",
      form: {
        title: "申請前置交易暫時豁免",
        failedRule: "觸發阻擋的規則 (規則類型 ID)",
        orderRef: "委託單關聯編號",
        portfolioId: "投資組合 ID",
        contractId: "合約 ID",
        reason: "豁免原因主題",
        justification: "業務合規說明理由與相關附件說明",
        expiry: "豁免截止日期",
        approver: "覆核審批人員",
        ack: "我已暸解相關風險，並同意系統記錄此申請操作歷程。",
        attachmentNote:
          "此模擬測試版本尚不支援附件上傳功能。",
        submitDisabled:
          "無法送出，後端目前不支援豁免申請 API。",
      },
      timeline: {
        title: "申請狀態與變更歷史",
        notRequested: "尚未發起申請",
        draft: "豁免申請草稿",
        pending: "合規審查中",
        approved: "豁免申請已核准生效",
        rejected: "申請已被審批人拒絕",
        expired: "豁免已過期",
      },
    },

    // Phase 2 — Audit Trail
    audit: {
      title: "合規審計歷史追蹤",
      description:
        "提供合規事件檢核的不可篡改日誌。因後端缺少 /compliance/audit 端點 (#12)，此處透過 API /compliance/checks/{id} 查詢對應 Check Group 的歷史檢核內容。",
      lookup: {
        title: "依檢查群組 ID 查詢",
        groupIdLabel: "檢查群組 ID (UUID)",
        cta: "查詢資料",
        loading: "查詢中…",
        empty: "請輸入上方合規記錄或前置交易檢核中的 check_group_id UUID 以載入詳情。",
      },
      records: "檢核記錄",
      breaches: "合規違規佐證資料",
      export: {
        label: "導出為 CSV 檔",
        disabled:
          "目前後端不支援導出 CSV 格式資料。",
      },
    },

    // Phase 2 — Permission Matrix
    permissions: {
      title: "個人帳戶合規權限矩陣",
      description:
        "展示合規功能模組與 RBAC 功能權限代碼對照表。數據來源為您的 PermissionsResp.functions[] 生效資料。",
      youHave: "您已具備此操作權限。",
      youDontHave: "您尚未獲得此權限項目。",
      action: "模組權限 / 操作功能",
      code: "權限項目編號",
      held: "具備該權限",
      table: {
        viewRules: "檢視合規規則清單",
        createRule: "建立規則實例項目",
        editBinding: "修改規則綁定配置 (Binding)",
        overrideBreach: "後置交易違規手動跳過 (Override)",
        adminRuleType: "管理及配置規則類型的基本屬性",
        runPreTrade: "即時執行 Pre-Trade / Post-Trade 檢查",
        viewExceptions: "檢視前置交易豁免申請 (規劃中)",
        requestException: "發起前置交易暫時豁免申請 (規劃中)",
        approveException: "核准或否決前置交易豁免申請 (規劃中)",
      },
    },

    // Investment decision embed
    decision: {
      title: "投資決策",
      description:
        "輸入擬議之委託單資訊並執行前置交易檢查。唯有在合規檢核通過後，該決策才允許送交後續處理。",
      sectionGuide: "決策流程說明",
      step1: "1. 填寫上方的擬議委託單細節",
      step2: "2. 執行合規檢查評估",
      step3:
        "3. 若結果為被阻擋 (BLOCKED)，請調整委託內容或申請豁免；若為通過 (PASS)，則此決策已就緒並等待審批。",
      readyForApproval: "合規檢查通過 — 此決策已準備好提交審審核。",
      blockedFromApproval:
        "合規檢查阻擋了此決策。請在提交審批前，修正所有合規違規事項。",
      warnFromApproval:
        "合規檢查提示警告訊息。在填寫風險確認說明後仍可提交審核。",
      submitForApprovalDisabled:
        "決策提交流程端點目前尚未與 API 進行完整整合。",
      submitForApproval: "提交此投資決策進行審批",
      ownerLabel: "決策執行人員",
      breakdownLabel: "投資比重摘要",
      notConfiguredTitle: "決策工作流尚未配置",
      notConfiguredCopy:
        "後端暫無投資決策生命周期的 API,因此本頁無法執行提交、覆核或審批操作。",
      notConfiguredAlternative:
        "若需對真實投組執行交易前合規模擬,請使用「Pre-trade Simulator」;若 workflow 已允許,可於該基金的 Holdings 頁面下單。",
    },

    // Badges
    badges: {
      verdict: {
        PASS: "通過",
        WARN: "警告",
        BLOCK: "攔截",
      },
      severity: {
        BLOCK: "攔截阻擋 (Blocker)",
        WARN: "警告提示 (Warning)",
        REQUIRE_APPROVAL: "需要審批覆核",
        MONITOR: "僅進行記錄監控",
      },
      status: {
        ACTIVE: "生效中",
        SCHEDULED: "排程生效中",
        EXPIRED: "已失效過期",
        DISABLED: "停用中",
      },
    },
  },
};
