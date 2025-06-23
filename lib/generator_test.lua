
local dir = 0.0
local rectAng = 3.1415 / 2

for i = 0,num-1,1 do
    local ang = 2 * (rnd() - 0.5) * bendMax
    local dirNext = dir+ang
    if math.abs(dirNext) > rectAng then
        ang = -ang
    end
    setAngle(i,ang)
	dir = dirNext
end
