import React, { useState } from 'react';
import { Link } from 'react-router-dom';

// Has Top level button that is clickable
// On hover, top level button style changes and dropdown opens
// On hovering dropdown buttons, style changes
// When hover ends, revert styles and destroy dropdown

export function DropdownTitle({ children, to, active }: { children: React.ReactNode, to: string, active: boolean }) {
  const count = React.Children.count(children);

  if (count !== 1) {
    throw new Error('DropdownTitle expects exactly one child.');
  }

  let className = 'dropdownTitle'
  if (active) {
    className += " active"
  }
  
  return (
    <Link className={className} to={to}><span>{children}</span></Link>
  )
}

export function DropdownItem({ children, to }: { children: React.ReactNode, to: string }) {
  const count = React.Children.count(children);

  if (count !== 1) {
    throw new Error('DropdownItem expects exactly one child.');
  }

  return (
    <li className='dropdownItem'><Link to={to}><span>{children}</span></Link></li>
  )
}

export function Dropdown({ title, titleTo, children }: { title: string, titleTo: string, children: React.ReactNode }) {
  const [dropdownActive, setDropdownActive] = useState(false)

  return (
    <div className="dropdownContainer" onMouseOver={() => setDropdownActive(true)} onMouseOut={() => setDropdownActive(false)}>
      <DropdownTitle active={dropdownActive} to={titleTo}>{title}</DropdownTitle>
      <div className="dropdownContent">
        <ul>
          {dropdownActive ? children : <></>}
        </ul>
      </div>
    </div>
  )
}
