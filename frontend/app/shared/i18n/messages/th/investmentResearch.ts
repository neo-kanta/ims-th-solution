import type { enInvestmentResearchMessages } from "../en/investmentResearch";

export const thInvestmentResearchMessages = {
  investmentResearch: {
    // Pages / Headers
    title: "บทวิเคราะห์การลงทุน",
    newReport: "สร้างบทวิเคราะห์ใหม่",
    editReport: "แก้ไขบทวิเคราะห์การลงทุน",
    researchReport: "บทวิเคราะห์การลงทุน",
    descriptionList: "บทวิเคราะห์การลงทุนโดยนักวิเคราะห์ พร้อมคำแนะนำและการติดตามสถานะวงจรชีวิต",
    descriptionNew: "บันทึกคำแนะนำของนักวิเคราะห์ในแบบร่างบทวิเคราะห์",
    loading: "กำลังโหลดบทวิเคราะห์การลงทุน…",
    loadingList: "กำลังโหลดบทวิเคราะห์การลงทุน…",
    untitledReport: "บทวิเคราะห์ไม่มีชื่อ",
    identification: "การระบุตัวตน",
    reportContent: "เนื้อหารายงาน",

    // Labels & Table columns / Form Fields
    reportNo: "เลขที่รายงาน",
    reportDate: "วันที่รายงาน",
    reportDateFrom: "จากวันที่รายงาน",
    reportDateTo: "ถึงวันที่รายงาน",
    effectiveDate: "วันที่มีผล",
    instrument: "ตราสาร",
    instrumentCode: "รหัสตราสาร",
    instrumentName: "ชื่อตราสาร",
    instrumentType: "ประเภทตราสาร",
    market: "ตลาด",
    currency: "สกุลเงิน",
    recommendation: "คำแนะนำ",
    reportStatus: "สถานะรายงาน",
    review: "การรีวิว",
    reviewStatus: "สถานะการรีวิว",
    titleField: "หัวข้อ",
    ownerUserId: "รหัสผู้ใช้เจ้าของ",
    authorUserId: "รหัสผู้ใช้ผู้เขียน",
    applicableContractId: "รหัสสัญญาที่เกี่ยวข้อง",
    applicableContract: "สัญญาที่เกี่ยวข้อง",
    ownerAuthor: "เจ้าของ / ผู้เขียน",
    typeMarketCurrency: "ประเภท / ตลาด / สกุลเงิน",
    companyOverview: "ภาพรวมบริษัท",
    companyOutlook: "แนวโน้มบริษัท",
    financialStatus: "สถานะทางการเงิน",
    esgComment: "ความคิดเห็น ESG",
    investmentAnalysis: "การวิเคราะห์การลงทุน",
    lifecycleNotes: "บันทึกวงจรชีวิต",
    rejectionReason: "เหตุผลการปฏิเสธ",
    postSubmissionNote: "บันทึกหลังการส่ง",
    created: "สร้างเมื่อ",
    lastUpdated: "แก้ไขล่าสุด",
    timezone: "เขตเวลา",
    timezoneValue: "เอเชีย/กรุงเทพฯ (UTC+7)",
    by: "โดย",
    keepCurrent: "คงเดิม",

    // Filters
    searchPlaceholder: "เลขที่รายงาน, รหัสย่อ, หัวข้อ",
    search: "ค้นหา",
    all: "ทั้งหมด",
    selectPlaceholder: "เลือก…",

    // Form help text and errors
    fieldLimits: {
      maxLength: "สูงสุด {max} ตัวอักษร",
      minLength: "บทวิเคราะห์การลงทุนต้องมีความยาวอย่างน้อย {min} ตัวอักษร",
      uuid: "ต้องอยู่ในรูปแบบ UUID",
      uuidOptional: "ต้องอยู่ในรูปแบบ UUID (หรือปล่อยว่าง)",
      currencyCode: "ต้องเป็นตัวอักษรภาษาอังกฤษตัวพิมพ์ใหญ่ 3 ตัว (เช่น THB)",
    },
    required: "จำเป็นต้องกรอก",
    characterCounter: "{count} / {max}",
    characterCounterAnalysis: "{count} / {min} ตัวอักษร",
    minimumAnalysisPlaceholder: "ขั้นต่ำ {min} ตัวอักษร",
    uuidOptionalPlaceholder: "UUID (ไม่บังคับ)",

    // Actions & Buttons
    actions: {
      apply: "นำไปใช้",
      reset: "รีเซ็ต",
      back: "ย้อนกลับ",
      edit: "แก้ไข",
      submit: "ส่ง",
      cancelSubmission: "ยกเลิกการส่ง",
      delete: "ลบ",
      cancel: "ยกเลิก",
      createReport: "สร้างรายงาน",
      saveChanges: "บันทึกการเปลี่ยนแปลง",
      prev: "ก่อนหน้า",
      next: "ถัดไป",
      shortcutHintCreate: "กด ⌘/Ctrl + Enter เพื่อสร้าง",
      shortcutHintSave: "กด ⌘/Ctrl + Enter เพื่อบันทึก",
    },

    // Confirm dialogs
    confirm: {
      delete: {
        title: "ลบบทวิเคราะห์การลงทุนนี้หรือไม่?",
        description: "การลบรายงานแบบไม่ถาวร (Soft-delete) จะนำรายงานออกจากรายการของนักวิเคราะห์และคิวการรีวิว แต่อ้างอิงการตรวจสอบ (audit trail) จะยังคงถูกเก็บรักษาไว้ ซึ่งไม่สามารถย้อนคืนได้จากหน้าเว็บ",
        confirmLabel: "ลบรายงาน",
      },
      submit: {
        title: "ส่งรายงานเพื่อรับการรีวิวหรือไม่?",
        description: "รายงานจะถูกเปลี่ยนสถานะจาก 'ยังไม่ส่ง' เป็น 'ส่งแล้ว' และจะปรากฏให้กลุ่มผู้รีวิวเห็น คุณสามารถยกเลิกการส่งรายงานนี้ได้จนกว่าการรีวิวจะเสร็จสมบูรณ์",
        confirmLabel: "ส่ง",
      },
      cancelSubmit: {
        title: "ยกเลิกการส่งหรือไม่?",
        description: "รายงานจะกลับไปสู่สถานะ 'ยังไม่ส่ง' และจะหายไปจากคิวการรีวิว ความเห็นที่มีอยู่เดิมของผู้รีวิวจะยังคงถูกเก็บรักษาไว้",
        confirmLabel: "ยกเลิกการส่ง",
      },
    },

    // Statuses
    status: {
      draft: "แบบร่าง",
      active: "ใช้งาน",
      expired: "หมดอายุ",
      rejected: "ปฏิเสธ",
    },
    reviewStatusValues: {
      notSubmitted: "ยังไม่ส่ง",
      submitted: "ส่งแล้ว",
      reviewCompleted: "รีวิวเสร็จสิ้น",
    },
    recommendationValues: {
      buy: "ซื้อ",
      sell: "ขาย",
      hold: "ถือ",
    },

    // Errors
    errors: {
      loadList: "โหลดบทวิเคราะห์การลงทุนไม่สำเร็จ",
      loadDetail: "โหลดบทวิเคราะห์การลงทุนไม่สำเร็จ",
      create: "สร้างบทวิเคราะห์การลงทุนไม่สำเร็จ",
      update: "อัปเดตบทวิเคราะห์การลงทุนไม่สำเร็จ",
      delete: "ลบบทวิเคราะห์การลงทุนไม่สำเร็จ",
      submit: "ส่งบทวิเคราะห์การลงทุนไม่สำเร็จ",
      cancelSubmit: "ยกเลิกการส่งบทวิเคราะห์การลงทุนไม่สำเร็จ",
    },

    // Empty state
    empty: {
      title: "ยังไม่มีบทวิเคราะห์การลงทุน",
      description: "สร้างแบบร่างเพื่อเริ่มต้นติดตามคำแนะนำของนักวิเคราะห์",
    },

    // Footer paging
    pager: {
      range: "หน้า {page} · แสดง {count} จากทั้งหมด {total}",
    },
  },
} as const;
