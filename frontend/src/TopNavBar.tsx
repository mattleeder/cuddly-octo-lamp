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
        <a href='/' className="siteName">BurrChess</a>

        <Dropdown title='Play' titleHref='/play'>
          <DropdownItem href="#">Test</DropdownItem>
        </Dropdown>

        <Dropdown title='Watch' titleHref='/watch'>
          <DropdownItem href="/watch">All</DropdownItem>
          <DropdownItem href="/watch?timeFormat=bullet">Bullet</DropdownItem>
          <DropdownItem href="/watch?timeFormat=blitz">Blitz</DropdownItem>
          <DropdownItem href="/watch?timeFormat=rapid">Rapid</DropdownItem>
          <DropdownItem href="/watch?timeFormat=classical">Classical</DropdownItem>
        </Dropdown>
      </div>

      <div className='navBarContainer right'>
        <NavbarSearch />

        <a href='#'>Sign In</a>

        <ToggleDropdown title={<Settings className='settingsCog'/>} >
          <ToggleDropdownSubmenu title="Theme">
            <ToggleDropdownItem href="#" onClick={() => console.log("Setting Theme: ThemeOne")}>ThemeOne</ToggleDropdownItem>
            <ToggleDropdownItem href="#" onClick={() => console.log("Setting Theme: ThemeOne")}>ThemeTwo</ToggleDropdownItem>
          </ToggleDropdownSubmenu>
          <ToggleDropdownSubmenu title="Language">
            <ToggleDropdownItem href="#" onClick={() => console.log("Setting Language: English")}>English</ToggleDropdownItem>
            <ToggleDropdownItem href="#" onClick={() => console.log("Setting Language: 中文")}>中文</ToggleDropdownItem>
          </ToggleDropdownSubmenu>
          <ToggleDropdownItem href="#"><LatencyDisplay /></ToggleDropdownItem>
        </ToggleDropdown>
      </div>
    </nav>
  )
}