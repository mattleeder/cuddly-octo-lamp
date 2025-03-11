import React from 'react';
import { Dropdown, DropdownItem } from "./Dropdown.tsx"
import { ToggleDropdown, ToggleDropdownItem, ToggleDropdownSubmenu } from './ToggleDropdown.tsx';
import { Settings } from 'lucide-react';
import { NavbarSearch } from './NavSearch.tsx';
import { LatencyDisplay } from './LatencyDisplay.tsx';

export function TopNavBar() {
  return (
    <nav className='topNavBar'>
      <div className='navBarContainer left'>
        <a href='#'>BurrChess</a>

        <Dropdown title='Play' titleHref='#'>
          <DropdownItem href="#">Test</DropdownItem>
        </Dropdown>

        <Dropdown title='Watch' titleHref='#'>
          <DropdownItem href="#">All</DropdownItem>
          <DropdownItem href="#">Bullet</DropdownItem>
          <DropdownItem href="#">Blitz</DropdownItem>
          <DropdownItem href="#">Rapid</DropdownItem>
          <DropdownItem href="#">Classical</DropdownItem>
        </Dropdown>
      </div>

      <div className='navBarContainer right'>
        <NavbarSearch />

        <a href='#'>Sign In</a>

        <ToggleDropdown title={<Settings />} style={{minWidth: "200px"}}>
          <ToggleDropdownSubmenu title="Theme">
            <ToggleDropdownItem href="#">ThemeOne</ToggleDropdownItem>
            <ToggleDropdownItem href="#">ThemeTwo</ToggleDropdownItem>
          </ToggleDropdownSubmenu>
          <ToggleDropdownSubmenu title="Language">
            <ToggleDropdownItem href="#">English</ToggleDropdownItem>
            <ToggleDropdownItem href="#">中文</ToggleDropdownItem>
          </ToggleDropdownSubmenu>
          <ToggleDropdownItem href="#"><LatencyDisplay /></ToggleDropdownItem>
        </ToggleDropdown>
      </div>
    </nav>
  )
}