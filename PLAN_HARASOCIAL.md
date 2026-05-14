# Kế Hoạch Tái Cấu Trúc Fly-Box (Harasocial Clone)

Mục tiêu cốt lõi: Biến Fly-Box thành một trung tâm quản lý tin nhắn và thông báo đa kênh, loại bỏ các rào cản đăng nhập không cần thiết trong giai đoạn phát triển, tập trung vào trải nghiệm người dùng cốt lõi (Core UX) giống như HaraSocial.

## Giai đoạn 1: Gỡ bỏ rào cản Authentication (Bypass Auth)
Mục đích: Không bắt buộc đăng nhập, không check JWT để quá trình code giao diện và test API diễn ra nhanh nhất.

1. **Backend**: 
   - Vô hiệu hóa JWT Middleware (`middlewares.AuthRequired()`).
   - Thay thế bằng một "Dummy Middleware": Luôn tự động gắn `UserID = 1` và `Role = "admin"` vào mọi request. Các API cũ vẫn hoạt động bình thường mà không cần token.
2. **Frontend**:
   - Tắt logic bảo vệ route ở `MainAppShell`.
   - Cập nhật `useAuthStore` để mặc định trạng thái luôn là `isAuthenticated: true` với một tài khoản ảo mặc định.
   - Ẩn/xóa trang `/login`. Mở web lên là vào thẳng `/inbox`.

## Giai đoạn 2: Xây dựng trang "Hub Kết Nối Kênh" (Channel Connection Hub)
Mục đích: Replicate giao diện kết nối kênh quản lý từ hình ảnh HaraSocial bạn cung cấp.

1. **Giao diện**:
   - Xây dựng trang `/hub/channels` (hoặc đặt làm trang chủ mặc định nếu chưa có kênh nào kết nối).
   - Thiết kế các Card kết nối trực quan cho:
     - **Facebook**
     - **Instagram**
     - **Zalo OA (đã xác thực và trả phí)**
     - **Shopee**
     - **TikTok for Business**
   - Mỗi Card sẽ có nút "Thêm kết nối" (màu xanh dương giống Harasocial) và link hướng dẫn.
2. **Logic Integration**:
   - Nút "Thêm kết nối" sẽ kích hoạt các luồng OAuth tương ứng đã có sẵn (hoặc giả lập thành công nếu chưa có API thật).

## Giai đoạn 3: Hoàn thiện Centralized Inbox (Hộp thư trung tâm)
Mục đích: Giao diện Chat phải phân biệt rõ ràng tin nhắn đến từ nguồn nào và quản lý tập trung.

1. **Sidebar Inbox**:
   - Thêm bộ lọc (Filter) mạnh mẽ: Lọc theo kênh (chỉ xem Facebook, chỉ xem Shopee...).
   - Hiển thị logo mini của nền tảng đè lên Avatar khách hàng để dễ nhận biết.
2. **Message Panel**:
   - Khu vực trả lời tin nhắn cần thay đổi linh hoạt các công cụ tùy theo nền tảng (ví dụ: Zalo có gửi sticker, Shopee có gửi sản phẩm).
   - Tích hợp tính năng gửi ảnh/file chung.

## Giai đoạn 4: Quản lý Thông báo tập trung (Notification Center)
Mục đích: Không chỉ tin nhắn, các thông báo về đơn hàng (Shopee, TikTok) hay tương tác (Facebook, Instagram) đều đổ về một chỗ.

1. **Notification Feed**:
   - Xây dựng một luồng (feed) thông báo realtime qua WebSocket.
   - Phân loại thông báo: Đơn hàng mới, Đánh giá (Review), Nhắc nhở.
