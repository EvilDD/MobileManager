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

// 保存剪贴板监听器ID，用于停止监听
let clipboardMonitoringInterval: number | null = null;
let lastClipboardText = '';
let isPasteListenerActive = false;

// 开始监听剪贴板
function startClipboardMonitoring(client: StreamClientScrcpy): void {
    if (clipboardMonitoringInterval !== null) {
        return; // 已经在监听中
    }

    console.log('开始监听PC剪贴板变化...');
    // 尝试获取初始剪贴板内容
    try {
        navigator.clipboard
            .readText()
            .then((text) => {
                console.log('成功读取初始剪贴板内容:', text ? text.substring(0, 20) + '...' : '(空)');
                lastClipboardText = text;
            })
            .catch((err) => {
                console.error('无法读取剪贴板(权限问题):', err);
                // 如果读取失败，改为监听粘贴事件
                startPasteEventListening(client);
            });
    } catch (e) {
        console.error('剪贴板API不可用:', e);
        // 如果API不可用，改为监听粘贴事件
        startPasteEventListening(client);
        return;
    }

    // 每秒检查一次剪贴板内容
    clipboardMonitoringInterval = window.setInterval(() => {
        try {
            navigator.clipboard
                .readText()
                .then((text) => {
                    // 只有当剪贴板内容变化时才发送
                    if (text && text !== lastClipboardText) {
                        lastClipboardText = text;
                        console.log('检测到剪贴板变化，同步内容:', text.substring(0, 20) + '...');
                        client.sendMessage(CommandControlMessage.createSetClipboardCommand(text));
                    }
                })
                .catch((err) => {
                    console.error('读取剪贴板失败(可能需要用户交互):', err);
                    // 如果持续失败，改为监听粘贴事件
                    if (clipboardMonitoringInterval !== null) {
                        window.clearInterval(clipboardMonitoringInterval);
                        clipboardMonitoringInterval = null;
                        startPasteEventListening(client);
                    }
                });
        } catch (e) {
            console.error('检查剪贴板时出错:', e);
            // 如果持续出错，改为监听粘贴事件
            if (clipboardMonitoringInterval !== null) {
                window.clearInterval(clipboardMonitoringInterval);
                clipboardMonitoringInterval = null;
                startPasteEventListening(client);
            }
        }
    }, 1000);
}

// 开始监听粘贴事件（作为备选方案）
function startPasteEventListening(client: StreamClientScrcpy): void {
    if (isPasteListenerActive) {
        return;
    }

    console.log('改为监听粘贴事件(Ctrl+V)...');
    isPasteListenerActive = true;

    // 添加粘贴事件监听
    document.addEventListener('paste', (e) => {
        if (!isPasteListenerActive) return;
        console.log('检测到粘贴事件!');
        const text = e.clipboardData?.getData('text');
        if (text) {
            console.log('从粘贴事件获取文本，发送到设备:', text.substring(0, 20) + (text.length > 20 ? '...' : ''));
            client.sendMessage(CommandControlMessage.createSetClipboardCommand(text));
            // 仍然记录最后的文本以便跟踪
            lastClipboardText = text;
            console.log('已同步粘贴内容到设备');
        } else {
            console.log('粘贴事件不包含文本内容');
        }
    });

    // // 显示提示
    // const notice = document.createElement('div');
    // notice.textContent = '请在此页面使用Ctrl+V粘贴以同步到设备';
    // notice.style.position = 'fixed';
    // notice.style.bottom = '10px';
    // notice.style.left = '10px';
    // notice.style.backgroundColor = 'rgba(0,0,0,0.7)';
    // notice.style.color = 'white';
    // notice.style.padding = '5px 10px';
    // notice.style.borderRadius = '3px';
    // notice.style.zIndex = '9999';
    // notice.style.fontSize = '12px';
    // document.body.appendChild(notice);

    // // 5秒后移除提示
    // setTimeout(() => {
    //     document.body.removeChild(notice);
    // }, 5000);
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

        // 自动启动剪贴板监听（而不是使用按钮控制）
        setTimeout(() => {
            console.log('自动启动PC剪贴板监听');
            startClipboardMonitoring(client);
        }, 1000);
        return new GoogToolBox(elements);
    }
}
