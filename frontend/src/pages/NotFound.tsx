import React, { useEffect, useRef } from 'react';
import DinoGame from 'react-chrome-dino-ts';
import 'react-chrome-dino-ts/index.css';
import '../styles/NotFound.css';

const NotFound = () => {
  const gameRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    // 移动端触摸支持
    const handleTouchStart = (e: TouchEvent) => {
      e.preventDefault();
      // 模拟空格键按下
      const spaceEvent = new KeyboardEvent('keydown', {
        key: ' ',
        code: 'Space',
        keyCode: 32,
        which: 32,
        bubbles: true,
        cancelable: true
      });
      document.dispatchEvent(spaceEvent);
    };

    const handleTouchEnd = (e: TouchEvent) => {
      e.preventDefault();
      // 模拟空格键释放
      const spaceEvent = new KeyboardEvent('keyup', {
        key: ' ',
        code: 'Space',
        keyCode: 32,
        which: 32,
        bubbles: true,
        cancelable: true
      });
      document.dispatchEvent(spaceEvent);
    };

    const handleClick = (e: MouseEvent) => {
      e.preventDefault();
      const spaceEvent = new KeyboardEvent('keydown', {
        key: ' ',
        code: 'Space',
        keyCode: 32,
        which: 32,
        bubbles: true,
        cancelable: true
      });
      document.dispatchEvent(spaceEvent);

      // 短暂延迟后释放按键
      setTimeout(() => {
        const spaceUpEvent = new KeyboardEvent('keyup', {
          key: ' ',
          code: 'Space',
          keyCode: 32,
          which: 32,
          bubbles: true,
          cancelable: true
        });
        document.dispatchEvent(spaceUpEvent);
      }, 100);
    };

    // 为游戏容器添加触摸事件监听
    const gameContainer = gameRef.current;
    if (gameContainer) {
      gameContainer.addEventListener('touchstart', handleTouchStart, { passive: false });
      gameContainer.addEventListener('touchend', handleTouchEnd, { passive: false });
      gameContainer.addEventListener('click', handleClick);
    }

    return () => {
      if (gameContainer) {
        gameContainer.removeEventListener('touchstart', handleTouchStart);
        gameContainer.removeEventListener('touchend', handleTouchEnd);
        gameContainer.removeEventListener('click', handleClick);
      }
    };
  }, []);

  return (
    <div className="not-found-container">
      <h1>404 - 页面未找到</h1>
      <p>哎呀！您要查找的页面不存在。不过别灰心，玩个游戏放松一下吧。</p>
      <p className="game-hint">💡 点击屏幕或按空格键控制恐龙跳跃</p>
      <div className="dino-game-container" ref={gameRef}>
        <DinoGame />
      </div>
    </div>
  );
};

export default NotFound;
