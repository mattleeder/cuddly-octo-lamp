import React from 'react';
import { Dropdown, DropdownItem } from "./Dropdown.tsx"
import { ToggleDropdown, ToggleDropdownItem, ToggleDropdownSubmenu } from './ToggleDropdown.tsx';
import { Settings } from 'lucide-react';
import { NavbarSearch } from './NavSearch.tsx';
import { LatencyDisplay } from './LatencyDisplay.tsx';
import { Link } from "react-router-dom";

export function TopNavBar() {
  return (
    <nav className='topNavBar'>
      <div className='navBarContainer left'>
        <Link to='/' className="siteName">BurrChess</Link>

        <Dropdown title='Play' titleTo='/play'>
          <DropdownItem to="#">Test</DropdownItem>
        </Dropdown>

        <Dropdown title='Watch' titleTo='/watch'>
          <DropdownItem to="/watch">All</DropdownItem>
          <DropdownItem to="/watch?timeFormat=bullet">Bullet</DropdownItem>
          <DropdownItem to="/watch?timeFormat=blitz">Blitz</DropdownItem>
          <DropdownItem to="/watch?timeFormat=rapid">Rapid</DropdownItem>
          <DropdownItem to="/watch?timeFormat=classical">Classical</DropdownItem>
        </Dropdown>
      </div>

      <div className='navBarContainer right'>
        <NavbarSearch />

        <Link to='/login'>Sign In</Link>

        <ToggleDropdown title={<Settings className='settingsCog'/>} >
          <ToggleDropdownSubmenu title="Theme">
            <ToggleDropdownItem to="#" onClick={() => console.log("Setting Theme: ThemeOne")}>ThemeOne</ToggleDropdownItem>
            <ToggleDropdownItem to="#" onClick={() => console.log("Setting Theme: ThemeOne")}>ThemeTwo</ToggleDropdownItem>
          </ToggleDropdownSubmenu>
          <ToggleDropdownSubmenu title="Language">
            <ToggleDropdownItem to="#" onClick={() => console.log("Setting Language: English")}>English</ToggleDropdownItem>
            <ToggleDropdownItem to="#" onClick={() => console.log("Setting Language: 中文")}>中文</ToggleDropdownItem>
          </ToggleDropdownSubmenu>
          <ToggleDropdownItem to="#"><LatencyDisplay /></ToggleDropdownItem>
        </ToggleDropdown>
      </div>
    </nav>
  )
}