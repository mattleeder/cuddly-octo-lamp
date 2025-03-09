import React, { useEffect, useRef, useState } from 'react';

// Has Top level button that is clickable
// On hover, top level button style changes and dropdown opens
// On hovering dropdown buttons, style changes
// When hover ends, revert styles and destroy dropdown

export function ToggleDropdownTitle({ children, active, onClick }: { children: React.ReactNode, active: boolean, onClick: () => void }) {
  const count = React.Children.count(children);

  if (count !== 1) {
    throw new Error('DropdownTitle expects exactly one child.');
  }

  let className = 'dropdownTitle'
  if (active) {
    className += " active"
  }
  
  return (
    <button className={className} onClick={() => onClick()}><span>{children}</span></button>
  )
}

export function ToggleDropdownItem({ children, href }: { children: React.ReactNode, href: string }) {
  const count = React.Children.count(children);

  if (count !== 1) {
    throw new Error('DropdownItem expects exactly one child.');
  }

  return (
    <li className='dropdownItem'><a href={href}><span>{children}</span></a></li>
  )
}

export function ToggleDropdown({ title, children }: { title: React.ReactNode, children: React.ReactNode }) {
  const [dropdownActive, setDropdownActive] = useState(false)
  const mouseOver = useRef(false)

  // Destroy dropdown if mouse not over
  useEffect(() => {
    window.addEventListener("click", () => {
      if (mouseOver.current != true) {
        setDropdownActive(false)
      }
    })
    return () => {
      window.removeEventListener("click", () => {
        if (!mouseOver.current != true) {
          setDropdownActive(false)
        }
      })
    }
  }, [])

  return (
    <div 
      className="dropdownContainer" 
      onMouseOver={() => {mouseOver.current = true}} 
      onMouseOut={() => {mouseOver.current = false}}
    >
      <ToggleDropdownTitle active={dropdownActive} onClick={() => {setDropdownActive(!dropdownActive)}}>{title}</ToggleDropdownTitle>
      <div className="dropdownContent">
        <ul>
          {dropdownActive ? children : <></>}
        </ul>
      </div>
    </div>
  )
}
