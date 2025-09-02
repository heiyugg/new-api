import React from 'react';

const NodelocIcon = ({ style = {}, ...props }) => {
  return (
    <svg
      width="20"
      height="20"
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      style={{ ...style }}
      {...props}
    >
      {/* Nodeloc logo - simplified network/node icon */}
      <circle cx="12" cy="12" r="3" fill="currentColor" />
      <circle cx="6" cy="6" r="2" fill="currentColor" />
      <circle cx="18" cy="6" r="2" fill="currentColor" />
      <circle cx="6" cy="18" r="2" fill="currentColor" />
      <circle cx="18" cy="18" r="2" fill="currentColor" />
      <line x1="9" y1="9" x2="15" y2="15" stroke="currentColor" strokeWidth="2" />
      <line x1="15" y1="9" x2="9" y2="15" stroke="currentColor" strokeWidth="2" />
      <line x1="8" y1="6" x2="10" y2="10" stroke="currentColor" strokeWidth="1.5" />
      <line x1="16" y1="6" x2="14" y2="10" stroke="currentColor" strokeWidth="1.5" />
      <line x1="8" y1="18" x2="10" y2="14" stroke="currentColor" strokeWidth="1.5" />
      <line x1="16" y1="18" x2="14" y2="14" stroke="currentColor" strokeWidth="1.5" />
    </svg>
  );
};

export default NodelocIcon;