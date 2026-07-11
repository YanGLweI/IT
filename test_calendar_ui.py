#!/usr/bin/env python3
"""通过浏览器UI测试日历通知功能"""
from playwright.sync_api import sync_playwright
import time

def run_test():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context(ignore_https_errors=True, viewport={"width": 1400, "height": 900})
        page = context.new_page()
        
        # 1. 打开登录页
        print("=== 1. 打开登录页 ===")
        page.goto('http://localhost:8081')
        page.wait_for_load_state('networkidle')
        page.screenshot(path='/tmp/test_01_login_page.png')
        print("  登录页已加载")
        
        # 2. 用ylw登 录
        print("=== 2. ylw登录 ===")
        page.fill('input[placeholder*="域账号"]', 'ylw')
        page.fill('input[placeholder*="密码"]', '!Qw2!Qw2')
        page.screenshot(path='/tmp/test_02_login_filled.png')
        page.click('button:has-text("登 录")')
        page.wait_for_load_state('networkidle')
        time.sleep(2)
        page.screenshot(path='/tmp/test_03_after_login.png')
        print(f"  登录后URL: {page.url}")
        
        # 3. 进入日历模块
        print("=== 3. 进入日历模块 ===")
        # 查找日历菜单
        calendar_link = page.locator('text=日历').first
        if calendar_link.is_visible():
            calendar_link.click()
            page.wait_for_load_state('networkidle')
            time.sleep(1)
            page.screenshot(path='/tmp/test_04_calendar_view.png')
            print("  已进入日历模块")
        else:
            # 尝试侧边栏菜单
            page.click('.el-menu-item:has-text("日历")')
            page.wait_for_load_state('networkidle')
            time.sleep(1)
            page.screenshot(path='/tmp/test_04_calendar_view.png')
            print("  通过侧边栏进入日历模块")
        
        # 4. 创建新日程
        print("=== 4. 创建新日程 ===")
        # 查找创建按钮
        create_btn = page.locator('button:has-text("新建")').first
        if not create_btn.is_visible():
            create_btn = page.locator('button:has-text("创建")').first
        if not create_btn.is_visible():
            create_btn = page.locator('.el-button--primary').first
        
        create_btn.click()
        time.sleep(1)
        page.screenshot(path='/tmp/test_05_create_dialog.png')
        print("  创建对话框已打开")
        
        # 填写标题
        title_input = page.locator('input[placeholder*="标题"]').first
        if not title_input.is_visible():
            title_input = page.locator('.el-dialog input').first
        title_input.fill('测试通知日程')
        
        # 设置时间 - 选择明天
        page.screenshot(path='/tmp/test_06_title_filled.png')
        
        # 添加参与者 zbj
        print("  添加参与者zbj...")
        # 查找参与者选择器
        participant_input = page.locator('input[placeholder*="参与"]').first
        if not participant_input.is_visible():
            participant_input = page.locator('input[placeholder*="搜索"]').first
        if participant_input.is_visible():
            participant_input.click()
            time.sleep(0.5)
            participant_input.fill('zbj')
            time.sleep(1)
            page.screenshot(path='/tmp/test_07_participant_search.png')
            # 选择zbj
            zbj_option = page.locator('text=zbj').first
            if zbj_option.is_visible():
                zbj_option.click()
                time.sleep(0.5)
                print("  已选择zbj")
        
        page.screenshot(path='/tmp/test_08_before_submit.png')
        
        # 提交创建
        submit_btn = page.locator('.el-dialog button:has-text("确定")').first
        if not submit_btn.is_visible():
            submit_btn = page.locator('.el-dialog .el-button--primary').first
        submit_btn.click()
        time.sleep(2)
        page.screenshot(path='/tmp/test_09_after_create.png')
        print("  日程创建完成")
        
        # 5. 退出登 录
        print("=== 5. 退出登录 ===")
        # 查找用户头像或退出按钮
        page.locator('.el-dropdown-link').first.click()
        time.sleep(0.5)
        page.screenshot(path='/tmp/test_10_user_menu.png')
        logout_btn = page.locator('text=退出').first
        if not logout_btn.is_visible():
            logout_btn = page.locator('text=注销').first
        logout_btn.click()
        page.wait_for_load_state('networkidle')
        time.sleep(1)
        page.screenshot(path='/tmp/test_11_logged_out.png')
        print("  已退出登录")
        
        # 6. 用zbj登 录
        print("=== 6. zbj登录 ===")
        page.fill('input[placeholder*="域账号"]', 'zbj')
        page.fill('input[placeholder*="密码"]', 'admin123456')
        page.click('button:has-text("登 录")')
        page.wait_for_load_state('networkidle')
        time.sleep(3)
        page.screenshot(path='/tmp/test_12_zbj_login.png')
        print(f"  zbj登录后URL: {page.url}")
        
        # 7. 检查通知铃铛
        print("=== 7. 检查通知 ===")
        # 查看是否有通知弹窗
        time.sleep(2)
        page.screenshot(path='/tmp/test_13_zbj_notifications_popup.png')
        
        # 点击铃铛图标
        bell = page.locator('.bell-icon').first
        if bell.is_visible():
            bell.click()
            time.sleep(1)
            page.screenshot(path='/tmp/test_14_zbj_notification_panel.png')
            print("  已打开通知面板")
            # 检查通知内容
            notif_items = page.locator('.notification-item').all()
            print(f"  通知数量: {len(notif_items)}")
            for i, item in enumerate(notif_items):
                text = item.inner_text()
                print(f"  通知{i+1}: {text[:50]}")
        else:
            page.screenshot(path='/tmp/test_14_no_bell.png')
            print("  未找到铃铛图标")
        
        # 8. 进入日历模块检查日程可见性
        print("=== 8. 检查日程可见性 ===")
        calendar_link = page.locator('text=日历').first
        if calendar_link.is_visible():
            calendar_link.click()
        else:
            page.click('.el-menu-item:has-text("日历")')
        page.wait_for_load_state('networkidle')
        time.sleep(2)
        page.screenshot(path='/tmp/test_15_zbj_calendar.png')
        
        # 检查是否能看到创建的日程
        page_content = page.content()
        has_event = '测试通知日程' in page_content
        print(f"  zbj能看到日程: {has_event}")
        
        # 最终截图
        page.screenshot(path='/tmp/test_16_final.png', full_page=True)
        
        print("\n=== 测试截图已保存 ===")
        for i in range(1, 17):
            print(f"  /tmp/test_{i:02d}_*.png")
        
        browser.close()

if __name__ == "__main__":
    run_test()
