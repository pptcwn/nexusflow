# NexusFlow - Zero-Redirect Review / Standard Routing Upgrade Checklist (2026)

**เป้าหมายหลัก**: อัพเกรดเป็น **Zero-Redirect Transparent Review / Standard Traffic Routing**  
**หลักการ Zero Redirect**: Browser ต้องไม่เห็น HTTP 3xx redirect; Cloudflare Worker เป็นคน fetch `REVIEW_ORIGIN` หรือ `STANDARD_ORIGIN` ภายใน edge แล้วส่ง response กลับ  
**วันที่เริ่มต้นโครงการ**: ____________________  
**ผู้รับผิดชอบ**: ____________________

---

## Phase 1: Foundation & Zero-Redirect Transparent Routing

**ระยะเวลา**: 3–5 วัน **ความสำคัญ**: สูงสุด

### งานที่ต้องทำ
- [ ] Worker และ Go Backend สื่อสารกันได้ดี
- [ ] ตั้งค่า Cloudflare WAF Rules แบบเต็ม (Block Bot, Rate Limit, Threat Score)
- [ ] เพิ่ม fallback เมื่อ Backend ล่ม โดย route ไป `review`
- [ ] เพิ่ม Error Handling และ Logging พื้นฐาน
- [ ] Deploy เวอร์ชันแรกขึ้น Production
- [ ] ยืนยันว่า Worker ไม่ส่ง HTTP 301 / 302 / 307 / 308 กลับไปที่ Browser
- [ ] ทดสอบ routing ไป `standard` กับ User ปกติ
- [ ] ทดสอบกับ Bot / curl / Headless Browser

**สถานะปัจจุบัน**:
- [ ] ยังไม่เริ่ม
- [ ] กำลังดำเนินการ
- [ ] เสร็จสมบูรณ์

**วันที่เริ่ม**: ________ **วันที่เสร็จ**: ________

**หมายเหตุ / ปัญหาที่พบ**:
> 

---

## Phase 2: Smart Decision Engine

**ระยะเวลา**: 4–6 วัน **ความสำคัญ**: สูง

### งานที่ต้องทำ
- [ ] ปรับโครงสร้าง `Decision` ให้มี `Confidence` (0.0–1.0)
- [ ] ปรับโครงสร้าง `Decision` ให้มี `AllowProgressive` เป็น compatibility field หรือพิจารณา deprecate เมื่อไม่ใช้ progressive routing แล้ว
- [ ] รวมสัญญาณหลายตัว (Fingerprint + Behavior + Referer + IP Reputation)
- [ ] เชื่อมต่อ Redis เพื่อเก็บประวัติผู้ใช้ (ถ้าต้องการ)
- [ ] ทดสอบ Decision Logic กับหลายสถานการณ์
- [ ] ให้ Backend ส่ง `mode` เป็น `review` หรือ `standard` พร้อม `confidence`

**สถานะปัจจุบัน**:
- [ ] ยังไม่เริ่ม
- [ ] กำลังดำเนินการ
- [ ] เสร็จสมบูรณ์

**วันที่เริ่ม**: ________ **วันที่เสร็จ**: ________

**หมายเหตุ / ปัญหาที่พบ**:
> 

---

## Phase 3: Advanced Behavior Tracking

**ระยะเวลา**: 3–5 วัน **ความสำคัญ**: สูง

### งานที่ต้องทำ
- [ ] พัฒนา `fingerprint.js` แบบเต็ม
- [ ] รวบรวมข้อมูลพฤติกรรม: watchTime, scroll, mouse movement, clicks, time on page, interaction
- [ ] ส่งข้อมูลไป Backend แบบ interval (ทุก 3 วินาที)
- [ ] ปรับ Backend ให้รับและคำนวณ Behavior Score
- [ ] รวม Behavior Score เข้ากับ Decision Logic
- [ ] ทดสอบกับ Browser จริง (Desktop + Mobile)

**สถานะปัจจุบัน**:
- [ ] ยังไม่เริ่ม
- [ ] กำลังดำเนินการ
- [ ] เสร็จสมบูรณ์

**วันที่เริ่ม**: ________ **วันที่เสร็จ**: ________

**หมายเหตุ / ปัญหาที่พบ**:
> 

---

## Phase 4: Review / Standard Origin Routing ★ สำคัญที่สุด

**ระยะเวลา**: 5–7 วัน **ความสำคัญ**: สูงสุด

### งานที่ต้องทำ
- [ ] Worker route traffic ไป `REVIEW_ORIGIN` หรือ `STANDARD_ORIGIN` ตามผลจาก Backend แบบ zero redirect
- [ ] Worker ใช้ `fetch()` ไปยัง origin ที่เลือก และส่ง response กลับโดยไม่เปลี่ยน URL ฝั่ง Browser
- [ ] ถ้า Backend ล่ม / response ผิดรูปแบบ / WAF signal เสี่ยง → route ไป `REVIEW_ORIGIN`
- [ ] ถ้า `mode = standard` และ confidence ผ่านเกณฑ์ → route ไป `STANDARD_ORIGIN`
- [ ] ตรวจ response status ว่าไม่ใช่ redirect status: 301, 302, 303, 307, 308
- [ ] ไม่ inject หรือสลับ content แบบซ่อนเร้นหลังโหลดหน้า
- [ ] ไม่ใช้คำหรือ config เก่าที่สื่อถึงการซ่อนหรือสลับ content แบบไม่โปร่งใส
- [ ] ทดสอบกับ User จริงว่าได้ `standard`
- [ ] ทดสอบกับ Bot / curl / Headless Browser ว่าได้ `review`

**สถานะปัจจุบัน**:
- [ ] ยังไม่เริ่ม
- [ ] กำลังดำเนินการ
- [ ] เสร็จสมบูรณ์

**วันที่เริ่ม**: ________ **วันที่เสร็จ**: ________

**หมายเหตุ / ปัญหาที่พบ**:
> 

---

## Phase 5: Monitoring, WAF & Maintenance

**ระยะเวลา**: ต่อเนื่อง **ความสำคัญ**: ปานกลาง

### งานที่ต้องทำ
- [ ] มี Logging พื้นฐานสำหรับ decision event (`review` / `standard`) โดยไม่เก็บข้อมูลต้องห้าม
- [ ] สร้าง Dashboard แบบง่ายเพื่อดูสถิติ
- [ ] มี Kill Switch ที่ใช้งานง่าย
- [ ] ตั้งแผนอัพเดท aggregate Fingerprint + WAF Rules ทุก 7–14 วัน
- [ ] มีระบบแจ้งเตือนเมื่อเกิดปัญหา
- [ ] เอกสารและโค้ดเป็นระเบียบ

**สถานะปัจจุบัน**:
- [ ] ยังไม่เริ่ม
- [ ] กำลังดำเนินการ
- [ ] เสร็จสมบูรณ์

**วันที่เริ่ม**: ________ **วันที่เสร็จ**: ________

**หมายเหตุ / ปัญหาที่พบ**:
> 

---

## สรุปความคืบหน้า

| Phase | สถานะ          | วันที่เริ่ม | วันที่เสร็จ | หมายเหตุ |
|-------|----------------|-------------|-------------|----------|
| 1     |                |             |             |          |
| 2     |                |             |             |          |
| 3     |                |             |             |          |
| 4     |                |             |             |          |
| 5     |                |             |             |          |

---

**คำแนะนำการใช้งาน**:
- ใช้ Checklist นี้ติดตามความคืบหน้าทุกวัน
- แนะนำทำตามลำดับ: **Phase 1 → Phase 2+3 → Phase 4 → Phase 5**
- อัพเดทสถานะและหมายเหตุทุกครั้งที่มีความคืบหน้า
