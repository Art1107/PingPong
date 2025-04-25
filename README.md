# PingPong Game (Player A vs Player B)

ระบบจำลองการแข่ง PingPong ผ่าน gRPC พร้อมบันทึกผลการแข่งขันลง MySQL

## ก่อนเริ่ม

1. **ติดตั้ง Proto Compiler**
   - ตามลิงค์ YouTube นี้:  
     [ติดตั้ง Protobuf](https://www.youtube.com/watch?v=ES_GI-lmhEU)

2. **สร้างฐานข้อมูล MySQL**
   - ชื่อฐานข้อมูล: `pingpong`
   - ไม่ต้องสร้างตาราง (สร้างให้อัตโนมัติตอนรัน)

## วิธีใช้

1. เปิดTerminal ที่โฟลเดอร์ `pingpong`
2. รันคำสั่ง:

   ```bash
   make run
