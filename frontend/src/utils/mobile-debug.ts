// 移动端调试工具
export const MobileDebug = {
    // 检测是否为移动设备
    isMobile: () => {
        return /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent);
    },

    // 检测是否为触摸设备
    isTouchDevice: () => {
        return 'ontouchstart' in window || navigator.maxTouchPoints > 0;
    },

    // 检测触摸事件支持
    checkTouchSupport: () => {
        const events = ['touchstart', 'touchmove', 'touchend'];
        const results: Record<string, boolean> = {};

        events.forEach(event => {
            results[event] = 'on' + event in window || event in document;
        });

        return results;
    },

    // 记录触摸事件
    logTouchEvents: () => {
        if (MobileDebug.isTouchDevice()) {
            console.log('Touch device detected');
            console.log('Touch support:', MobileDebug.checkTouchSupport());

            // 添加全局触摸事件监听器
            document.addEventListener('touchstart', (e) => {
                console.log('Touch start:', e.target);
            }, { passive: true });

            document.addEventListener('touchend', (e) => {
                console.log('Touch end:', e.target);
            }, { passive: true });
        }
    },

    // 测试按钮点击
    testButtonClicks: () => {
        const buttons = document.querySelectorAll('.reaction-button');
        buttons.forEach((button, index) => {
            button.addEventListener('click', () => {
                console.log(`Button ${index + 1} clicked`);
            });

            button.addEventListener('touchstart', () => {
                console.log(`Button ${index + 1} touch start`);
            });

            button.addEventListener('touchend', () => {
                console.log(`Button ${index + 1} touch end`);
            });
        });
    },

    // 检查按钮大小
    checkButtonSizes: () => {
        const buttons = document.querySelectorAll('.reaction-button');
        buttons.forEach((button, index) => {
            const rect = button.getBoundingClientRect();
            console.log(`Button ${index + 1} size:`, {
                width: rect.width,
                height: rect.height,
                meetsTouchTarget: rect.width >= 44 && rect.height >= 44
            });
        });
    }
};

// 自动启用调试
if (import.meta.env.DEV) {
    window.addEventListener('load', () => {
        MobileDebug.logTouchEvents();
        setTimeout(MobileDebug.testButtonClicks, 1000);
        setTimeout(MobileDebug.checkButtonSizes, 1500);
    });
}