export const thPlaceholderMessages = {
  placeholders: {
    approvalWorkflow: {
      title: "เวิร์กโฟลว์การอนุมัติ",
      description: "เครื่องมือเวิร์กโฟลว์การอนุมัติกำลังจะมา",
    },
    approvalConfiguration: {
      title: "การตั้งค่าการอนุมัติ",
      description: "หน้าตั้งค่าการอนุมัติกำลังจะมา",
    },
    analysisReports: {
      title: "รายงานการวิเคราะห์",
      description: "รายงานการวิเคราะห์กำลังจะมา",
    },
    investmentDecisions: {
      title: "การตัดสินใจลงทุน",
      description: "เครื่องมือการตัดสินใจลงทุนกำลังจะมา",
    },
    executionOrders: {
      title: "การดำเนินการและคำสั่งซื้อขาย",
      description: "บันทึกรายการและตรวจสอบสถานะการดำเนินการ",
      notConfiguredTitle: "การดำเนินการบริหารต่อกองทุน",
      notConfiguredCopy:
        "ระบบยังไม่มี API บริหารคำสั่งซื้อขายแบบรวมศูนย์ การบันทึกรายการซื้อ/ขายจริงต้องทำที่แท็บ Operation ของกองทุน โดยแบ็กเอนด์จะตรวจสอบ Pre-trade Compliance ซ้ำในขั้นตอนบันทึก หากบันทึกสำเร็จแปลว่าผ่านเงื่อนไขและถูกบันทึกใน Ledger เรียบร้อยแล้ว",
    },
    investmentReview: {
      title: "ทบทวนการลงทุน",
      description: "เครื่องมือทบทวนการลงทุนกำลังจะมา",
    },
    leaveDelegation: {
      title: "ลาและมอบหมายงานแทน",
      description: "เครื่องมือลาและมอบหมายงานแทนกำลังจะมา",
    },
    agentManagement: {
      title: "จัดการตัวแทน",
      description: "เครื่องมือจัดการตัวแทนกำลังจะมา",
    },
    accountManagement: {
      title: "จัดการบัญชี",
      description: "เครื่องมือจัดการบัญชีกำลังจะมา",
    },
    groupManagement: {
      title: "จัดการกลุ่ม",
      description: "เครื่องมือจัดการกลุ่มกำลังจะมา",
    },
    accountLog: {
      title: "บันทึกบัญชี",
      description: "เครื่องมือบันทึกบัญชีกำลังจะมา",
    },
    auditLog: {
      title: "บันทึกการตรวจสอบ",
      description: "เครื่องมือบันทึกการตรวจสอบกำลังจะมา",
    },
    notificationSettings: {
      title: "การตั้งค่าการแจ้งเตือน",
      description: "การตั้งค่าการแจ้งเตือนกำลังจะมา",
    },
  },
} as const;
