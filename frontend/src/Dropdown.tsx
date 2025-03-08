import React, { useState } from 'react';

// Has Top level button that is clickable
// On hover, top level button style changes and dropdown opens
// On hovering dropdown buttons, style changes
// When hover ends, revert styles and destroy dropdown

export function DropdownTitle({ children, href, active }: { children: React.ReactNode, href: string, active: boolean }) {
  const count = React.Children.count(children);

  if (count !== 1) {
    throw new Error('DropdownTitle expects exactly one child.');
  }

  let className = 'dropdownTitle'
  if (active) {
    className += " active"
  }
  
  return (
    <li className={className}><a href={href}><span>{children}</span></a></li>
  )
}

export function DropdownItem({ children, href }: { children: React.ReactNode, href: string }) {
  const count = React.Children.count(children);

  if (count !== 1) {
    throw new Error('DropdownItem expects exactly one child.');
  }

  return (
    <li className='dropdownItem'><a href={href}><span>{children}</span></a></li>
  )
}

export function Dropdown({ title, titleHref, children }: { title: string, titleHref: string, children: React.ReactNode }) {
  const [dropdownActive, setDropdownActive] = useState(false)

  return (
    <div className="dropdownContainer" onMouseOver={() => setDropdownActive(true)} onMouseOut={() => setDropdownActive(false)}>
      <ul className="dropdownMenu">
        <DropdownTitle active={dropdownActive} href={titleHref}>{title}</DropdownTitle>
        {dropdownActive ? children : <></>}
      </ul>
    </div>
  )
}
