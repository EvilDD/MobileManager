import { ToolBox } from '../../toolbox/ToolBox';
import KeyEvent from '../android/KeyEvent';
import SvgImage from '../../ui/SvgImage';
import { KeyCodeControlMessage } from '../../controlMessage/KeyCodeControlMessage';
import { ToolBoxButton } from '../../toolbox/ToolBoxButton';
import { ToolBoxElement } from '../../toolbox/ToolBoxElement';
import { ToolBoxCheckbox } from '../../toolbox/ToolBoxCheckbox';
import { StreamClientScrcpy } from '../client/StreamClientScrcpy';
import { BasePlayer } from '../../player/BasePlayer';
import { CommandControlMessage } from '../../controlMessage/CommandControlMessage';

// 保存监听状态
let isPasteListenerActive = false;

// 开始监听剪贴板
function startClipboardMonitoring(client: StreamClientScrcpy): void {
    if (isPasteListenerActive) {
        return; // 已经在监听中
    }

    console.log('启动剪贴板监听 (隐形激活模式)...');

    // 创建隐形覆盖层以捕获首次交互
    const interactionOverlayId = 'invisible-interaction-layer';
    if (!document.getElementById(interactionOverlayId)) {
        const overlay = document.createElement('div');
        overlay.id = interactionOverlayId;
        overlay.style.position = 'absolute'; // 或 'fixed'，取决于布局
        overlay.style.top = '0';
        overlay.style.left = '0';
        overlay.style.width = '100%';
        overlay.style.height = '100%';
        overlay.style.zIndex = '5000'; // 确保在内容之上，但在控件之下（如果需要）
        overlay.style.opacity = '0'; // 完全透明
        overlay.style.cursor = 'default'; // 保持默认光标

        const activateListener = () => {
            console.log('隐形层被点击，移除并聚焦iframe');
            // 移除覆盖层
            if (overlay.parentNode) {
                overlay.parentNode.removeChild(overlay);
            }
            // 尝试聚焦iframe
            try {
                window.focus();
                // 如果有更具体的元素（如播放器容器）可以聚焦，效果可能更好
                // document.querySelector('.player-container')?.focus();
            } catch (e) {
                console.warn('尝试聚焦iframe失败:', e);
            }
            // 移除自身的监听器，确保只触发一次
            overlay.removeEventListener('click', activateListener);
            overlay.removeEventListener('touchstart', activateListener); // 兼容触摸设备
        };

        overlay.addEventListener('click', activateListener);
        overlay.addEventListener('touchstart', activateListener); // 兼容触摸设备

        document.body.appendChild(overlay);
        console.log('添加了隐形交互层');
    }

    // 直接启动粘贴事件监听
    document.addEventListener('paste', function pasteHandler(e) {
        console.log('检测到粘贴事件!');
        const text = e.clipboardData?.getData('text');
        if (text) {
            console.log('从粘贴事件获取文本，发送到设备:', text.substring(0, 20) + (text.length > 20 ? '...' : ''));
            client.sendMessage(CommandControlMessage.createSetClipboardCommand(text));
            console.log('已同步粘贴内容到设备');
        } else {
            console.log('粘贴事件不包含文本内容');
        }
    });
    isPasteListenerActive = true;
}

const BUTTONS = [
    {
        title: 'Power',
        code: KeyEvent.KEYCODE_POWER,
        icon: SvgImage.Icon.POWER,
    },
    {
        title: 'Volume up',
        code: KeyEvent.KEYCODE_VOLUME_UP,
        icon: SvgImage.Icon.VOLUME_UP,
    },
    {
        title: 'Volume down',
        code: KeyEvent.KEYCODE_VOLUME_DOWN,
        icon: SvgImage.Icon.VOLUME_DOWN,
    },
    {
        title: 'Back',
        code: KeyEvent.KEYCODE_BACK,
        icon: SvgImage.Icon.BACK,
    },
    {
        title: 'Home',
        code: KeyEvent.KEYCODE_HOME,
        icon: SvgImage.Icon.HOME,
    },
    {
        title: 'Overview',
        code: KeyEvent.KEYCODE_APP_SWITCH,
        icon: SvgImage.Icon.OVERVIEW,
    },
];

export class GoogToolBox extends ToolBox {
    protected constructor(list: ToolBoxElement<any>[]) {
        super(list);
    }

    public static createToolBox(
        udid: string,
        player: BasePlayer,
        client: StreamClientScrcpy,
        moreBox?: HTMLElement,
    ): GoogToolBox {
        const playerName = player.getName();
        const list = BUTTONS.slice();
        const handler = <K extends keyof HTMLElementEventMap, T extends HTMLElement>(
            type: K,
            element: ToolBoxElement<T>,
        ) => {
            if (!element.optional?.code) {
                return;
            }
            const { code } = element.optional;
            const action = type === 'mousedown' ? KeyEvent.ACTION_DOWN : KeyEvent.ACTION_UP;
            const event = new KeyCodeControlMessage(action, code, 0, 0);
            client.sendMessage(event);
        };
        const elements: ToolBoxElement<any>[] = list.map((item) => {
            const button = new ToolBoxButton(item.title, item.icon, {
                code: item.code,
            });
            button.addEventListener('mousedown', handler);
            button.addEventListener('mouseup', handler);
            return button;
        });
        if (player.supportsScreenshot) {
            const screenshot = new ToolBoxButton('Take screenshot', SvgImage.Icon.CAMERA);
            screenshot.addEventListener('click', () => {
                player.createScreenshot(client.getDeviceName());
            });
            elements.push(screenshot);
        }

        const keyboard = new ToolBoxCheckbox(
            'Capture keyboard',
            SvgImage.Icon.KEYBOARD,
            `capture_keyboard_${udid}_${playerName}`,
        );
        keyboard.addEventListener('click', (_, el) => {
            const element = el.getElement();
            client.setHandleKeyboardEvents(element.checked);
        });
        elements.push(keyboard);

        if (moreBox) {
            const displayId = player.getVideoSettings().displayId;
            const id = `show_more_${udid}_${playerName}_${displayId}`;
            const more = new ToolBoxCheckbox('More', SvgImage.Icon.MORE, id);
            more.addEventListener('click', (_, el) => {
                const element = el.getElement();
                moreBox.style.display = element.checked ? 'block' : 'none';
            });
            elements.push(more);
        }

        // 自动启动剪贴板监听（无提示模式）
        setTimeout(() => {
            startClipboardMonitoring(client);
        }, 1000);
        return new GoogToolBox(elements);
    }
}
