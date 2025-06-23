
local dir = 0.0
local rectAng = math.pi / 2

for i = 0,num-1,1 do
    local ang = 2 * (rnd() - 0.5) * bendMax
    if math.abs(dir+ang) > rectAng then
        ang = -ang
    end
    setAngle(i,ang)
	dir = dir+ang
end
