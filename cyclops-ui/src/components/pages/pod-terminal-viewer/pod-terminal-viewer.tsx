import {Terminal} from 'xterm';
import {FitAddon} from 'xterm-addon-fit';
import * as React from 'react';
import 'xterm/css/xterm.css';
import {useCallback, useEffect} from 'react';
import {debounceTime, takeUntil} from 'rxjs/operators';
import {fromEvent, ReplaySubject, Subject} from 'rxjs';
export interface PodTerminalViewerProps {
    applicationName: string;
    applicationNamespace: string;
    projectName: string;
    containerName: string;
    onClickContainer?: (group: any, i: number, tab: string) => any;
}
export interface ShellFrame {
    operation: string;
    data?: string;
    rows?: number;
    cols?: number;
}

export const PodTerminalViewer: React.FC<PodTerminalViewerProps> = ({
    applicationName,
    applicationNamespace,
    projectName,
    containerName,
    onClickContainer
}) => {
    const terminalRef = React.useRef(null);
    const fitAddon = new FitAddon();
    let terminal: Terminal;
    let webSocket: WebSocket;
    const keyEvent = new ReplaySubject<KeyboardEvent>(2);
    let connSubject = new ReplaySubject<ShellFrame>(100);
    let incommingMessage = new Subject<ShellFrame>();
    const unsubscribe = new Subject<void>();
    let connected = false;

    const onTerminalSendString = (str: string) => {
        if (connected) {
            webSocket.send(JSON.stringify({operation: 'stdin', data: str, rows: terminal.rows, cols: terminal.cols}));
        }
    };

    const onTerminalResize = () => {
        if (connected) {
            webSocket.send(
                JSON.stringify({
                    operation: 'resize',
                    cols: terminal.cols,
                    rows: terminal.rows
                })
            );
        }
    };

    const onConnectionMessage = (e: MessageEvent) => {
        const msg = JSON.parse(e.data);
        connSubject.next(msg);
    };

    const onConnectionOpen = () => {
        connected = true;
        onTerminalResize(); // fit the screen first time
        terminal.focus();
    };

    const onConnectionClose = () => {
        if (!connected) return;
        if (webSocket) webSocket.close();
        connected = false;
    };

    const handleConnectionMessage = (frame: ShellFrame) => {
        terminal.write("init session");
        incommingMessage.next(frame);
    };

    const disconnect = () => {
        if (webSocket) {
            webSocket.close();
        }

        if (connSubject) {
            connSubject.complete();
            connSubject = new ReplaySubject<ShellFrame>(100);
        }

        if (terminal) {
            terminal.dispose();
        }

        incommingMessage.complete();
        incommingMessage = new Subject<ShellFrame>();
    };

    function initTerminal(node: HTMLElement) {
        if (connSubject) {
            connSubject.complete();
            connSubject = new ReplaySubject<ShellFrame>(100);
        }

        if (terminal) {
            terminal.dispose();
        }

        terminal = new Terminal({
            convertEol: true,
            fontFamily: 'Menlo, Monaco, Courier New, monospace',
            fontSize: 14,
            fontWeight: 400,
            cursorBlink: true
        });
        terminal.options = {
            theme: {
                background: '#333'
            }
        };
        terminal.loadAddon(fitAddon);
        terminal.open(node);
        fitAddon.fit();

        connSubject.pipe(takeUntil(unsubscribe)).subscribe(frame => {
            handleConnectionMessage(frame);
        });

        terminal.onResize(onTerminalResize);
        terminal.onKey(key => {
            keyEvent.next(key.domEvent);
        });
        terminal.onData(onTerminalSendString);
    }

    function setupConnection() {
        const {name = '', namespace = ''} = {};
        const url = `${location.host}`.replace(/\/$/, '');
        webSocket = new WebSocket(
            `${
                location.protocol === 'https:' ? 'wss' : 'ws'
            }://${url}/terminal?pod=${name}&container=${containerName}&appName=${applicationName}&appNamespace=${applicationNamespace}&projectName=${projectName}&namespace=${namespace}`
        );
        webSocket.onopen = onConnectionOpen;
        webSocket.onclose = onConnectionClose;
        webSocket.onerror = e => {
            onConnectionClose();
        };
        webSocket.onmessage = onConnectionMessage;
    }

    useEffect(() => {
        const resizeHandler = fromEvent(window, 'resize')
            .pipe(debounceTime(1000))
            .subscribe(() => {
                if (fitAddon) {
                    fitAddon.fit();
                }
            });
        return () => {
            resizeHandler.unsubscribe(); // unsubscribe resize callback
            unsubscribe.next();
            unsubscribe.complete();

            // clear connection and close terminal
            if (webSocket) {
                webSocket.close();
            }

            if (connSubject) {
                connSubject.complete();
            }

            if (terminal) {
                terminal.dispose();
            }

            incommingMessage.complete();
        };
    }, [containerName]);

    return (
        <div className='row'/>
    );
};
