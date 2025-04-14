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

    console.log('启动剪贴板监听...');
    
    // 直接启动粘贴事件监听，移除复制事件监听（复制事件监听不稳定）
    startPasteEventListening(client);
    
    // 添加视觉提示，告诉用户首次需要在页面内使用Ctrl+V
    showClipboardHelpTip();
}

// 开始监听粘贴事件（作为主要方案）
function startPasteEventListening(client: StreamClientScrcpy): void {
    if (isPasteListenerActive) {
        return;
    }

    console.log('监听粘贴事件(Ctrl+V)...');
    isPasteListenerActive = true;

    // 添加粘贴事件监听
    document.addEventListener('paste', (e) => {
        if (!isPasteListenerActive) return;
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
}

// 显示帮助提示
function showClipboardHelpTip(): void {
    const tipId = 'clipboard-help-tip';
    
    // 如果已经存在提示，不再重复创建
    if (document.getElementById(tipId)) {
        return;
    }
    
    // 创建提示元素
    const container = document.createElement('div');
    container.id = tipId;
    container.style.position = 'fixed';
    container.style.bottom = '10px';
    container.style.left = '10px';
    container.style.backgroundColor = 'rgba(0,0,0,0.8)';
    container.style.color = 'white';
    container.style.padding = '12px 16px';
    container.style.borderRadius = '6px';
    container.style.zIndex = '9999';
    container.style.maxWidth = '90%';
    container.style.boxShadow = '0 4px 12px rgba(0,0,0,0.2)';
    container.style.fontFamily = 'Arial, sans-serif';
    
    // 添加标题
    const title = document.createElement('div');
    title.textContent = '如何将PC剪贴板内容发送到设备';
    title.style.fontWeight = 'bold';
    title.style.fontSize = '14px';
    title.style.marginBottom = '8px';
    
    // 添加说明
    const steps = document.createElement('div');
    steps.innerHTML = 
        '1. 在PC上复制需要的文本<br>' +
        '2. <b>在此页面内点击并按Ctrl+V</b><br>' +
        '3. 内容将自动同步到设备';
    steps.style.fontSize = '12px';
    steps.style.lineHeight = '1.5';
    
    // 添加关闭按钮
    const closeBtn = document.createElement('button');
    closeBtn.textContent = '了解';
    closeBtn.style.marginTop = '10px';
    closeBtn.style.padding = '4px 12px';
    closeBtn.style.border = 'none';
    closeBtn.style.borderRadius = '4px';
    closeBtn.style.backgroundColor = '#3f85ff';
    closeBtn.style.color = 'white';
    closeBtn.style.cursor = 'pointer';
    closeBtn.style.fontSize = '12px';
    closeBtn.style.fontWeight = 'bold';
    closeBtn.onclick = () => {
        container.style.display = 'none';
    };
    
    // 组装提示
    container.appendChild(title);
    container.appendChild(steps);
    container.appendChild(closeBtn);
    document.body.appendChild(container);
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
