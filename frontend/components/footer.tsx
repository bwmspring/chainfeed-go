import { Logo } from "./logo";

export function Footer() {
  return (
    <footer className="border-t border-slate-200 dark:border-slate-800 bg-white/80 dark:bg-slate-900/80 backdrop-blur-xl mt-20">
      <div className="container mx-auto px-4 py-12">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
          {/* 品牌信息 */}
          <div className="space-y-4">
            <div className="flex items-center gap-2">
              <Logo className="w-7 h-7" />
              <span className="text-xl font-bold bg-gradient-to-r from-purple-600 to-blue-600 bg-clip-text text-transparent">
                ChainFeed
              </span>
            </div>
            <p className="text-sm text-muted-foreground">
              像刷 Twitter 一样追踪链上活动
            </p>
            <div className="flex items-center gap-3">
              <a
                href="https://twitter.com/llxxs"
                target="_blank"
                rel="noopener noreferrer"
                className="text-2xl hover:scale-110 transition-transform"
                title="Twitter"
              >
                𝕏
              </a>
              <a
                href="https://github.com/bwmspring/chainfeed"
                target="_blank"
                rel="noopener noreferrer"
                className="text-2xl hover:scale-110 transition-transform"
                title="GitHub"
              >
                <svg aria-hidden="true" height="24" viewBox="0 0 24 24" version="1.1" width="24" data-view-component="true" className="fill-current">
                  <path d="M12 1C5.923 1 1 5.923 1 12c0 4.867 3.149 8.979 7.521 10.436.55.096.756-.233.756-.522 0-.262-.013-1.128-.013-2.049-2.764.509-3.479-.674-3.699-1.292-.124-.317-.66-1.293-1.127-1.554-.385-.207-.936-.715-.014-.729.866-.014 1.485.797 1.691 1.128.99 1.663 2.571 1.196 3.204.907.096-.715.385-1.196.701-1.471-2.448-.275-5.005-1.224-5.005-5.432 0-1.196.426-2.186 1.128-2.956-.111-.275-.496-1.402.11-2.915 0 0 .921-.288 3.024 1.128a10.193 10.193 0 0 1 2.75-.371c.936 0 1.871.123 2.75.371 2.104-1.43 3.025-1.128 3.025-1.128.605 1.513.221 2.64.111 2.915.701.77 1.127 1.747 1.127 2.956 0 4.222-2.571 5.157-5.019 5.432.399.344.743 1.004.743 2.035 0 1.471-.014 2.654-.014 3.025 0 .289.206.632.756.522C19.851 20.979 23 16.854 23 12c0-6.077-4.922-11-11-11Z"></path>
                </svg>
              </a>
            </div>
          </div>

          {/* 产品 */}
          <div>
            <h3 className="font-semibold mb-4">产品</h3>
            <ul className="space-y-2 text-sm text-muted-foreground">
              <li>
                <a href="/feed" className="hover:text-purple-600 transition-colors">
                  实时动态
                </a>
              </li>
              <li>
                <a href="#" className="hover:text-purple-600 transition-colors">
                  地址监控
                </a>
              </li>
              <li>
                <a href="#" className="hover:text-purple-600 transition-colors">
                  数据分析
                </a>
              </li>
              <li>
                <a href="#" className="hover:text-purple-600 transition-colors">
                  API 文档
                </a>
              </li>
            </ul>
          </div>

          {/* 资源 */}
          <div>
            <h3 className="font-semibold mb-4">资源</h3>
            <ul className="space-y-2 text-sm text-muted-foreground">
              <li>
                <a href="#" className="hover:text-purple-600 transition-colors">
                  使用指南
                </a>
              </li>
              <li>
                <a href="#" className="hover:text-purple-600 transition-colors">
                  开发文档
                </a>
              </li>
              <li>
                <a href="#" className="hover:text-purple-600 transition-colors">
                  博客
                </a>
              </li>
              <li>
                <a href="#" className="hover:text-purple-600 transition-colors">
                  更新日志
                </a>
              </li>
            </ul>
          </div>

          {/* 联系方式 */}
          <div>
            <h3 className="font-semibold mb-4">联系我们</h3>
            <ul className="space-y-2 text-sm text-muted-foreground">
              <li className="flex items-center gap-2">
                <span>📧</span>
                <a href="mailto:bwm029@gmail.com" className="hover:text-purple-600 transition-colors">
                  bwm029@gmail.com
                </a>
              </li>
              <li className="flex items-center gap-2">
                <span>💼</span>
                <a href="#" className="hover:text-purple-600 transition-colors">
                  商务合作
                </a>
              </li>
              <li className="flex items-center gap-2">
                <span>🐛</span>
                <a href="https://github.com/chainfeed/issues" className="hover:text-purple-600 transition-colors">
                  反馈问题
                </a>
              </li>
            </ul>
          </div>
        </div>

        {/* 友情链接 */}
        <div className="mt-12 pt-8 border-t border-slate-200 dark:border-slate-800">
          <div className="flex flex-wrap items-center justify-center gap-6 text-sm text-muted-foreground mb-6">
            <span className="font-medium">友情链接：</span>
            <a href="https://ethereum.org" target="_blank" rel="noopener noreferrer" className="hover:text-purple-600 transition-colors">
              Ethereum
            </a>
            <a href="https://etherscan.io" target="_blank" rel="noopener noreferrer" className="hover:text-purple-600 transition-colors">
              Etherscan
            </a>
            <a href="https://alchemy.com" target="_blank" rel="noopener noreferrer" className="hover:text-purple-600 transition-colors">
              Alchemy
            </a>
            <a href="https://rainbow.me" target="_blank" rel="noopener noreferrer" className="hover:text-purple-600 transition-colors">
              Rainbow
            </a>
          </div>

          {/* 版权信息 */}
          <div className="text-center text-sm text-muted-foreground">
            <p>© 2026 ChainFeed. All rights reserved.</p>
            <div className="flex items-center justify-center gap-4 mt-2">
              <a href="#" className="hover:text-purple-600 transition-colors">
                隐私政策
              </a>
              <span>·</span>
              <a href="#" className="hover:text-purple-600 transition-colors">
                服务条款
              </a>
              <span>·</span>
              <a href="#" className="hover:text-purple-600 transition-colors">
                Cookie 政策
              </a>
            </div>
          </div>
        </div>
      </div>
    </footer>
  );
}
