import "./scene.css";

const WAREHOUSES = [
  { wall: "#FFD66B", roof: "#FF8C69" },
  { wall: "#9FD8FF", roof: "#5E7CE2" },
  { wall: "#FFB3C7", roof: "#B56CD6" },
  { wall: "#B8F0C8", roof: "#2FA38A" },
];

function Warehouse({ wall, roof }: { wall: string; roof: string }) {
  return (
    <svg className="wh" width="169" height="100" viewBox="0 0 220 130" aria-hidden="true">
      <rect x="10" y="44" width="200" height="82" rx="6" fill={wall} />
      <path d="M2 52 Q110 -4 218 52 Z" fill={roof} />
      <rect x="22" y="70" width="34" height="22" rx="5" fill="#EAF8FF" />
      <rect x="164" y="70" width="34" height="22" rx="5" fill="#EAF8FF" />
      <rect x="72" y="72" width="76" height="54" rx="5" fill="#FFFFFF" />
      <g stroke="#C9D3EA" strokeWidth="3" strokeLinecap="round">
        <line x1="78" y1="84" x2="142" y2="84" />
        <line x1="78" y1="96" x2="142" y2="96" />
        <line x1="78" y1="108" x2="142" y2="108" />
      </g>
      <circle cx="96" cy="60" r="3.2" fill="#26305E" />
      <circle cx="124" cy="60" r="3.2" fill="#26305E" />
      <path d="M102 65 q8 6 16 0" fill="none" stroke="#26305E" strokeWidth="2.6" strokeLinecap="round" />
    </svg>
  );
}

function Truck({ trailer, cab }: { trailer: string; cab: string }) {
  return (
    <svg className="bob" viewBox="0 0 160 80" aria-hidden="true">
      <rect x="4" y="8" width="104" height="52" rx="9" fill={trailer} />
      <rect x="14" y="20" width="26" height="22" rx="5" fill="#FFFFFF" opacity="0.85" />
      <path d="M22 31 h10 M27 26 v10" stroke={trailer} strokeWidth="3" strokeLinecap="round" />
      <rect x="108" y="26" width="46" height="34" rx="9" fill={cab} />
      <rect x="124" y="31" width="25" height="18" rx="6" fill="#DFF4FF" />
      <circle cx="132" cy="38" r="2.2" fill="#26305E" />
      <circle cx="142" cy="38" r="2.2" fill="#26305E" />
      <path d="M133 43 q4.5 3.5 9 0" fill="none" stroke="#26305E" strokeWidth="2" strokeLinecap="round" />
      <circle cx="152" cy="52" r="3" fill="#FFE066" />
      <rect x="4" y="56" width="150" height="7" rx="3.5" fill="#26305E" />
      {[30, 84, 130].map((cx) => (
        <g key={cx}>
          <circle cx={cx} cy="64" r="12" fill="#26305E" />
          <g className="wheel">
            <circle cx={cx} cy="64" r="6" fill="#DDE5F5" />
            <line x1={cx - 5} y1="64" x2={cx + 5} y2="64" stroke="#26305E" strokeWidth="2" strokeLinecap="round" />
          </g>
        </g>
      ))}
    </svg>
  );
}

export default function Scene() {
  const half = Array.from({ length: 13 }, (_, i) => WAREHOUSES[i % WAREHOUSES.length]);

  return (
    <div className="scene" aria-hidden="true">
      <div className="sun" />
      <div className="cloud c1" />
      <div className="cloud c2" />
      <div className="cloud c3" />

      <div className="skyline">
        <div className="skyline-track">
          {[0, 1].map((k) => (
            <div className="skyline-half" key={k}>
              {half.map((w, i) => (
                <Warehouse key={i} wall={w.wall} roof={w.roof} />
              ))}
            </div>
          ))}
        </div>
      </div>

      <div className="grass" />
      <div className="road">
        <div className="road-line" />
      </div>

      <div className="truck truck-far t-far1">
        <Truck trailer="#7FB5FF" cab="#FFC93C" />
      </div>
      <div className="truck truck-far t-far2">
        <Truck trailer="#B8F0C8" cab="#FF8C69" />
      </div>
      <div className="truck truck-near t-near flip">
        <Truck trailer="#FFB3C7" cab="#5E7CE2" />
      </div>
    </div>
  );
}
