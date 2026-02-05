export function Logo({ className = "w-8 h-8" }: { className?: string }) {
  return (
    <svg
      className={className}
      viewBox="0 0 100 100"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      {/* 链条环 */}
      <circle
        cx="30"
        cy="40"
        r="18"
        stroke="url(#gradient1)"
        strokeWidth="6"
        fill="none"
      />
      <circle
        cx="70"
        cy="60"
        r="18"
        stroke="url(#gradient2)"
        strokeWidth="6"
        fill="none"
      />
      
      {/* RSS 波纹 */}
      <path
        d="M 20 70 Q 30 60, 40 70"
        stroke="url(#gradient3)"
        strokeWidth="4"
        strokeLinecap="round"
        fill="none"
      />
      <path
        d="M 15 75 Q 30 55, 45 75"
        stroke="url(#gradient3)"
        strokeWidth="3"
        strokeLinecap="round"
        fill="none"
        opacity="0.6"
      />
      
      {/* 渐变定义 */}
      <defs>
        <linearGradient id="gradient1" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stopColor="#3b82f6" />
          <stop offset="100%" stopColor="#06b6d4" />
        </linearGradient>
        <linearGradient id="gradient2" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stopColor="#06b6d4" />
          <stop offset="100%" stopColor="#8b5cf6" />
        </linearGradient>
        <linearGradient id="gradient3" x1="0%" y1="0%" x2="100%" y2="0%">
          <stop offset="0%" stopColor="#06b6d4" />
          <stop offset="100%" stopColor="#3b82f6" />
        </linearGradient>
      </defs>
    </svg>
  );
}
