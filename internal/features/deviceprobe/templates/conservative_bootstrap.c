#include "pd_api.h"
#include <string.h>

_Static_assert(sizeof(PDSystemEvent) <= 4, "PDSystemEvent must fit a 32-bit call slot");
_Static_assert(kEventMirrorEnded <= INT32_MAX, "PDSystemEvent values must fit int32_t");
_Static_assert(sizeof(uint32_t) == 4, "event argument must be 32-bit");
_Static_assert(sizeof(uintptr_t) == 4, "device pointers must be 32-bit");
_Static_assert(sizeof(int) == 4, "Playdate callback result must be 32-bit");
_Static_assert(sizeof(float) == 4, "Playdate float samples must be IEEE-754 binary32 slots");

extern void runtimeRun(void) __asm__("runtime.run");
extern uintptr_t runtimeStackTop __asm__("runtime.stackTop");
extern void* runtimeSCB __asm__("playdateRuntimeSCB");
extern int goEventHandler(PlaydateAPI*, PDSystemEvent, uint32_t);
extern int goUpdate(void);
static PlaydateAPI* activePlaydate;
extern void goMenuCallback(uint32_t id);
static void bridgeMenuCallback(void* userdata){goMenuCallback((uint32_t)(uintptr_t)userdata);}
extern void goSerialMessage(uintptr_t message);
extern void goScoreCallback(int32_t kind,uintptr_t score,uintptr_t error);
extern void goBoardsCallback(uintptr_t list,uintptr_t error);
extern void goScoresCallback(uintptr_t list,uintptr_t error);
static void bridgeSerialMessage(const char* message){goSerialMessage((uintptr_t)message);}
static void bridgeAddScoreCallback(PDScore* score,const char* error){goScoreCallback(0,(uintptr_t)score,(uintptr_t)error);if(score)activePlaydate->scoreboards->freeScore(score);}
static void bridgePersonalBestCallback(PDScore* score,const char* error){goScoreCallback(1,(uintptr_t)score,(uintptr_t)error);if(score)activePlaydate->scoreboards->freeScore(score);}
static void bridgeBoardsCallback(PDBoardsList* list,const char* error){goBoardsCallback((uintptr_t)list,(uintptr_t)error);if(list)activePlaydate->scoreboards->freeBoardsList(list);}
static void bridgeScoresCallback(PDScoresList* list,const char* error){goScoresCallback((uintptr_t)list,(uintptr_t)error);if(list)activePlaydate->scoreboards->freeScoresList(list);}

static int booted;
static uint32_t runtimeSCBShadow[2];
__attribute__((section(".bss.playdate_runtime_heap"), aligned(16), used))
unsigned char playdateRuntimeHeap[256 * 1024];
static int bridgeUpdate(void* userdata);

static void prepareRuntimeBoundary(void)
{
	register uintptr_t stackPointer __asm__("sp");
	runtimeStackTop = stackPointer;
}

int eventHandler(PlaydateAPI* playdate, PDSystemEvent event, uint32_t arg)
{
	int result;
    if (event == kEventInit && !booted) {
		activePlaydate = playdate;
		runtimeSCB = runtimeSCBShadow;
		prepareRuntimeBoundary();
        runtimeRun();
        booted = 1;
    }
	result = goEventHandler(playdate, event, arg);
	if (event == kEventInit && result == 0)
	{
		playdate->system->setUpdateCallback(bridgeUpdate, playdate);
		playdate->system->setSerialMessageCallback(bridgeSerialMessage);
		playdate->system->resetElapsedTime();
	}
	return result;
}

static int bridgeUpdate(void* userdata)
{
	(void)userdata;
	prepareRuntimeBoundary();
	return goUpdate();
}

void bridgeClear(void)
{
	activePlaydate->graphics->clear(kColorWhite);
}
void bridgeLog(const char* message){activePlaydate->system->logToConsole("%s",message);}
uintptr_t bridgeGetFrame(void){return (uintptr_t)activePlaydate->graphics->getFrame();}
void bridgeMarkUpdatedRows(int32_t start,int32_t end){activePlaydate->graphics->markUpdatedRows(start,end);}

void bridgeDrawText(const char* text, uintptr_t length, int32_t x, int32_t y)
{
	activePlaydate->graphics->drawText(text, length, kUTF8Encoding, x, y);
}
uintptr_t bridgeLoadFont(const char* path, uintptr_t* error) { return (uintptr_t)activePlaydate->graphics->loadFont(path, (const char**)error); }
void bridgeSetFont(uintptr_t font) { activePlaydate->graphics->setFont((LCDFont*)font); }
int32_t bridgeTextWidth(uintptr_t font, const char* text, uintptr_t length) { return activePlaydate->graphics->getTextWidth((LCDFont*)font,text,length,kUTF8Encoding,0); }
int32_t bridgeFontHeight(uintptr_t font) { return activePlaydate->graphics->getFontHeight((LCDFont*)font); }
void bridgeFreeFont(uintptr_t font) { activePlaydate->system->realloc((void*)font,0); }
void bridgeSetTextTracking(int32_t value){activePlaydate->graphics->setTextTracking(value);}
int32_t bridgeTextTracking(void){return activePlaydate->graphics->getTextTracking();}
void bridgeSetTextLeading(int32_t value){activePlaydate->graphics->setTextLeading(value);}
void bridgeDrawTextInRect(const char* text,uintptr_t length,int32_t x,int32_t y,int32_t width,int32_t height,int32_t wrap,int32_t align){activePlaydate->graphics->drawTextInRect(text,length,kUTF8Encoding,x,y,width,height,(PDTextWrappingMode)wrap,(PDTextAlignment)align);}
int32_t bridgeTextHeight(uintptr_t font,const char* text,uintptr_t length,int32_t width,int32_t wrap,int32_t tracking,int32_t leading){return activePlaydate->graphics->getTextHeightForMaxWidth((LCDFont*)font,text,length,width,kUTF8Encoding,(PDTextWrappingMode)wrap,tracking,leading);}
uintptr_t bridgeFontGlyph(uintptr_t font,uint32_t code,uint32_t next,uintptr_t* bitmap,int32_t* advance,int32_t* kerning){LCDFontPage* page=activePlaydate->graphics->getFontPage((LCDFont*)font,code);if(!page)return 0;LCDBitmap* image=0;LCDFontGlyph* glyph=activePlaydate->graphics->getPageGlyph(page,code,&image,(int*)advance);if(!glyph)return 0;*bitmap=(uintptr_t)image;*kerning=next?activePlaydate->graphics->getGlyphKerning(glyph,code,next):0;return(uintptr_t)glyph;}

uint32_t bridgeCurrentTimeMilliseconds(void)
{
	return activePlaydate->system->getCurrentTimeMilliseconds();
}

void bridgeExitToLauncher(void) { activePlaydate->system->exitToLauncher(); }
static uint32_t bridgeFloatBits(float value);
void bridgeSetAccelerometerEnabled(int32_t enabled){activePlaydate->system->setPeripheralsEnabled(enabled?kAccelerometer:kNone);}
void bridgeAccelerometer(float* x,float* y,float* z){activePlaydate->system->getAccelerometer(x,y,z);}
int32_t bridgePowerStatus(void){return activePlaydate->system->getPowerStatus();}
uint32_t bridgeBatteryPercentageBits(void){return bridgeFloatBits(activePlaydate->system->getBatteryPercentage());}
uint32_t bridgeBatteryVoltageBits(void){return bridgeFloatBits(activePlaydate->system->getBatteryVoltage());}
uint32_t bridgeSystemVolumeBits(void){return bridgeFloatBits(activePlaydate->system->getVolume());}
int32_t bridgeReduceFlashing(void){return activePlaydate->system->getReduceFlashing();}
int32_t bridgeTimezoneOffsetSeconds(void){return activePlaydate->system->getTimezoneOffset();}
int32_t bridgeUses24HourTime(void){return activePlaydate->system->shouldDisplay24HourTime();}
int32_t bridgeLanguage(void){return activePlaydate->system->getLanguage();}
uintptr_t bridgeLocalizedText(const char* key,int32_t language){return(uintptr_t)activePlaydate->system->getLocalizedText(key,(PDLanguage)language);}
void bridgeFree(uintptr_t pointer){activePlaydate->system->realloc((void*)pointer,0);}
uintptr_t bridgeAddMenuItem(const char* title,int32_t kind,const char** options,int32_t count,int32_t value,uint32_t callback){PDMenuItem* item;void* userdata=(void*)(uintptr_t)callback;if(kind==1)item=activePlaydate->system->addCheckmarkMenuItem(title,value,bridgeMenuCallback,userdata);else if(kind==2)item=activePlaydate->system->addOptionsMenuItem(title,options,count,bridgeMenuCallback,userdata);else item=activePlaydate->system->addMenuItem(title,bridgeMenuCallback,userdata);return(uintptr_t)item;}
int32_t bridgeAddScore(const char* board,uint32_t value){return activePlaydate->scoreboards->addScore(board,value,bridgeAddScoreCallback);}
int32_t bridgeGetPersonalBest(const char* board){return activePlaydate->scoreboards->getPersonalBest(board,bridgePersonalBestCallback);}
int32_t bridgeGetScoreboards(void){return activePlaydate->scoreboards->getScoreboards(bridgeBoardsCallback);}
int32_t bridgeGetScores(const char* board){return activePlaydate->scoreboards->getScores(board,bridgeScoresCallback);}
uint32_t bridgeScoreNumber(uintptr_t value,int32_t field){PDScore* score=(PDScore*)value;if(!score)return 0;return field==0?score->rank:score->value;}
uintptr_t bridgeScoreText(uintptr_t value,int32_t field){PDScore* score=(PDScore*)value;if(!score)return 0;return(uintptr_t)(field==0?score->player:score->boardID);}
uint32_t bridgeListNumber(uintptr_t value,int32_t field){if(!value)return 0;if(field==0)return((PDBoardsList*)value)->count;if(field==1)return((PDBoardsList*)value)->lastUpdated;PDScoresList* list=(PDScoresList*)value;return field==2?(uint32_t)list->playerIncluded:list->limit;}
uintptr_t bridgeListText(uintptr_t value,int32_t index,int32_t field){if(!value)return 0;if(index<0)return(uintptr_t)((PDScoresList*)value)->boardID;if(field==0)return(uintptr_t)((PDBoardsList*)value)->boards[index].boardID;if(field==1)return(uintptr_t)((PDBoardsList*)value)->boards[index].name;return(uintptr_t)((PDScoresList*)value)->scores[index].player;}
uint32_t bridgeListItemNumber(uintptr_t value,int32_t index,int32_t field){PDListScore* score=&((PDScoresList*)value)->scores[index];return field==0?score->rank:score->value;}
void bridgeRemoveMenuItem(uintptr_t item){activePlaydate->system->removeMenuItem((PDMenuItem*)item);}
uintptr_t bridgeMenuItemTitle(uintptr_t item){return(uintptr_t)activePlaydate->system->getMenuItemTitle((PDMenuItem*)item);}
void bridgeSetMenuItemTitle(uintptr_t item,const char* title){activePlaydate->system->setMenuItemTitle((PDMenuItem*)item,title);}
int32_t bridgeMenuItemValue(uintptr_t item){return activePlaydate->system->getMenuItemValue((PDMenuItem*)item);}
void bridgeSetMenuItemValue(uintptr_t item,int32_t value){activePlaydate->system->setMenuItemValue((PDMenuItem*)item,value);}
const char* bridgeFileError(void){return activePlaydate->file->geterr();}
uintptr_t bridgeFileOpen(const char* path,int32_t options){return(uintptr_t)activePlaydate->file->open(path,(FileOptions)options);}
int32_t bridgeFileClose(uintptr_t file){return activePlaydate->file->close((SDFile*)file);}
int32_t bridgeFileRead(uintptr_t file,void* buffer,uint32_t length){return activePlaydate->file->read((SDFile*)file,buffer,length);}
int32_t bridgeFileWrite(uintptr_t file,const void* buffer,uint32_t length){return activePlaydate->file->write((SDFile*)file,buffer,length);}
int32_t bridgeFileFlush(uintptr_t file){return activePlaydate->file->flush((SDFile*)file);}
int32_t bridgeFileTell(uintptr_t file){return activePlaydate->file->tell((SDFile*)file);}
int32_t bridgeFileSeek(uintptr_t file,int32_t position,int32_t whence){return activePlaydate->file->seek((SDFile*)file,position,whence);}
int32_t bridgeFileStat(const char* path,int32_t* values){FileStat value;int result=activePlaydate->file->stat(path,&value);if(result<0)return result;values[0]=value.isdir;values[1]=(int32_t)value.size;values[2]=value.m_year;values[3]=value.m_month;values[4]=value.m_day;values[5]=value.m_hour;values[6]=value.m_minute;values[7]=value.m_second;return 0;}
typedef struct{char** items;int32_t count;int32_t failed;}BridgeFileList;
static void bridgeCollectFile(const char* name,void* userdata){BridgeFileList* list=userdata;if(list->failed)return;char* copy=activePlaydate->system->realloc(NULL,strlen(name)+1);if(!copy){list->failed=1;return;}strcpy(copy,name);char** items=activePlaydate->system->realloc(list->items,sizeof(char*)*(list->count+1));if(!items){activePlaydate->system->realloc(copy,0);list->failed=1;return;}list->items=items;list->items[list->count++]=copy;}
uintptr_t bridgeFileList(const char* path,int32_t showHidden,int32_t* count,int32_t* result){BridgeFileList* list=activePlaydate->system->realloc(NULL,sizeof(BridgeFileList));if(!list){*count=0;*result=-2;return 0;}list->items=NULL;list->count=0;list->failed=0;*result=activePlaydate->file->listfiles(path,bridgeCollectFile,list,showHidden);if(list->failed)*result=-2;*count=list->count;return(uintptr_t)list;}
uintptr_t bridgeFileListItem(uintptr_t list,int32_t index){return(uintptr_t)((BridgeFileList*)list)->items[index];}
void bridgeFileListFree(uintptr_t value){BridgeFileList* list=(BridgeFileList*)value;if(!list)return;for(int i=0;i<list->count;i++)activePlaydate->system->realloc(list->items[i],0);activePlaydate->system->realloc(list->items,0);activePlaydate->system->realloc(list,0);}
int32_t bridgeFileMkdir(const char* path){return activePlaydate->file->mkdir(path);}
int32_t bridgeFileRemove(const char* path,int32_t recursive){return activePlaydate->file->unlink(path,recursive);}
int32_t bridgeFileRename(const char* from,const char* to){return activePlaydate->file->rename(from,to);}

uint32_t bridgeButtons(void) { PDButtons value = 0; activePlaydate->system->getButtonState(&value, NULL, NULL); return (uint32_t)value; }
static uint32_t bridgeFloatBits(float value) { union { float value; uint32_t bits; } conversion = { .value = value }; return conversion.bits; }
uint32_t bridgeCrankAngleBits(void) { return bridgeFloatBits(activePlaydate->system->getCrankAngle()); }
uint32_t bridgeCrankDeltaBits(void) { return bridgeFloatBits(activePlaydate->system->getCrankChange()); }
int32_t bridgeCrankDocked(void) { return activePlaydate->system->isCrankDocked(); }
uint32_t bridgeFrameDeltaBits(void) { float value = activePlaydate->system->getElapsedTime(); activePlaydate->system->resetElapsedTime(); return bridgeFloatBits(value); }
uintptr_t bridgeLoadBitmap(const char* path, const char** error) { return (uintptr_t)activePlaydate->graphics->loadBitmap(path, error); }
uintptr_t bridgeLoadBitmapTable(const char* path, const char** error) { return (uintptr_t)activePlaydate->graphics->loadBitmapTable(path, error); }
uintptr_t bridgeBitmapTableFrame(uintptr_t table, int32_t index) { return (uintptr_t)activePlaydate->graphics->getTableBitmap((LCDBitmapTable*)table, index); }
void bridgeFreeBitmapTable(uintptr_t table) { activePlaydate->graphics->freeBitmapTable((LCDBitmapTable*)table); }
uintptr_t bridgeNewBitmap(int32_t width, int32_t height) { return (uintptr_t)activePlaydate->graphics->newBitmap(width, height, kColorClear); }
void bridgeFreeBitmap(uintptr_t bitmap) { activePlaydate->graphics->freeBitmap((LCDBitmap*)bitmap); }
void bridgeBitmapSize(uintptr_t bitmap, int32_t* width, int32_t* height) { int nativeWidth, nativeHeight; activePlaydate->graphics->getBitmapData((LCDBitmap*)bitmap, &nativeWidth, &nativeHeight, NULL, NULL, NULL); *width = nativeWidth; *height = nativeHeight; }
static LCDColor bridgeBitmapColor(int32_t color) { return color == 1 ? kColorWhite : color == 2 ? kColorBlack : kColorClear; }
void bridgeFillBitmap(uintptr_t bitmap, int32_t color) { activePlaydate->graphics->clearBitmap((LCDBitmap*)bitmap, bridgeBitmapColor(color)); }
void bridgeBitmapData(uintptr_t bitmap,int32_t* width,int32_t* height,int32_t* rowbytes,uint8_t** mask,uint8_t** data){int w,h,r;activePlaydate->graphics->getBitmapData((LCDBitmap*)bitmap,&w,&h,&r,mask,data);*width=w;*height=h;*rowbytes=r;}
uintptr_t bridgeCopyBitmap(uintptr_t bitmap){return(uintptr_t)activePlaydate->graphics->copyBitmap((LCDBitmap*)bitmap);}
void bridgeLoadIntoBitmap(const char* path,uintptr_t bitmap,const char** error){activePlaydate->graphics->loadIntoBitmap(path,(LCDBitmap*)bitmap,error);}
uintptr_t bridgeNewBitmapTable(int32_t count,int32_t width,int32_t height){return(uintptr_t)activePlaydate->graphics->newBitmapTable(count,width,height);}
void bridgeLoadIntoBitmapTable(const char* path,uintptr_t table,const char** error){activePlaydate->graphics->loadIntoBitmapTable(path,(LCDBitmapTable*)table,error);}
int32_t bridgeSetBitmapMask(uintptr_t bitmap,uintptr_t mask){return activePlaydate->graphics->setBitmapMask((LCDBitmap*)bitmap,(LCDBitmap*)mask);}
uintptr_t bridgeGetBitmapMask(uintptr_t bitmap){return(uintptr_t)activePlaydate->graphics->getBitmapMask((LCDBitmap*)bitmap);}
int32_t bridgeCheckMaskCollision(uintptr_t a,int32_t ax,int32_t ay,int32_t af,uintptr_t b,int32_t bx,int32_t by,int32_t bf,int32_t x,int32_t y,int32_t width,int32_t height){return activePlaydate->graphics->checkMaskCollision((LCDBitmap*)a,ax,ay,(LCDBitmapFlip)af,(LCDBitmap*)b,bx,by,(LCDBitmapFlip)bf,LCDMakeRect(x,y,width,height));}
uintptr_t bridgeRotatedBitmapBits(uintptr_t bitmap,uint32_t degrees,uint32_t sx,uint32_t sy,int32_t* size){union{uint32_t bits;float value;}d={.bits=degrees},x={.bits=sx},y={.bits=sy};int value;LCDBitmap* result=activePlaydate->graphics->rotatedBitmap((LCDBitmap*)bitmap,d.value,x.value,y.value,&value);*size=value;return(uintptr_t)result;}
uintptr_t bridgeCopyDisplayBuffer(void){return(uintptr_t)activePlaydate->graphics->copyFrameBufferBitmap();}
void bridgeDrawBitmap(uintptr_t bitmap, int32_t x, int32_t y) { activePlaydate->graphics->drawBitmap((LCDBitmap*)bitmap, x, y, kBitmapUnflipped); }
void bridgeDrawScaledBitmapBits(uintptr_t bitmap, int32_t x, int32_t y, uint32_t scaleX, uint32_t scaleY) { union { uint32_t bits; float value; } sx = { .bits = scaleX }, sy = { .bits = scaleY }; activePlaydate->graphics->drawScaledBitmap((LCDBitmap*)bitmap, x, y, sx.value, sy.value); }
void bridgeDrawRotatedBitmapBits(uintptr_t bitmap,int32_t x,int32_t y,uint32_t degrees,uint32_t centerX,uint32_t centerY,uint32_t scaleX,uint32_t scaleY){union{uint32_t bits;float value;}d={.bits=degrees},cx={.bits=centerX},cy={.bits=centerY},sx={.bits=scaleX},sy={.bits=scaleY};activePlaydate->graphics->drawRotatedBitmap((LCDBitmap*)bitmap,x,y,d.value,cx.value,cy.value,sx.value,sy.value);}
static LCDColor bridgePrimitivePaint(int32_t solid, uintptr_t pattern, int32_t patterned) { if(patterned) return (LCDColor)pattern; return solid==1?kColorWhite:solid==2?kColorBlack:solid==3?kColorXOR:kColorClear; }
void bridgeDrawPrimitive(int32_t kind,int32_t x1,int32_t y1,int32_t x2,int32_t y2,int32_t x3,int32_t y3,int32_t lineWidth,int32_t solid,uint32_t startAngle,uint32_t endAngle,uintptr_t pattern,int32_t patterned) { union{uint32_t bits;float value;} start={.bits=startAngle},end={.bits=endAngle}; LCDColor color=bridgePrimitivePaint(solid,pattern,patterned); switch(kind){case 0:activePlaydate->graphics->drawLine(x1,y1,x2,y2,lineWidth,color);break;case 1:activePlaydate->graphics->drawRect(x1,y1,x2,y2,color);break;case 2:activePlaydate->graphics->fillRect(x1,y1,x2,y2,color);break;case 3:activePlaydate->graphics->drawEllipse(x1,y1,x2,y2,lineWidth,start.value,end.value,color);break;case 4:activePlaydate->graphics->fillEllipse(x1,y1,x2,y2,start.value,end.value,color);break;case 5:activePlaydate->graphics->fillTriangle(x1,y1,x2,y2,x3,y3,color);break;case 6:activePlaydate->graphics->drawLine(x1,y1,x2,y2,lineWidth,color);activePlaydate->graphics->drawLine(x2,y2,x3,y3,lineWidth,color);activePlaydate->graphics->drawLine(x3,y3,x1,y1,lineWidth,color);break;} }
void bridgeFillPolygon(int32_t count,uintptr_t points,int32_t rule,int32_t solid,uintptr_t pattern,int32_t patterned){activePlaydate->graphics->fillPolygon(count,(int*)points,bridgePrimitivePaint(solid,pattern,patterned),(LCDPolygonFillRule)rule);}
void bridgeDrawRoundedRect(int32_t x,int32_t y,int32_t width,int32_t height,int32_t radius,int32_t lineWidth,int32_t solid,uintptr_t pattern,int32_t patterned){activePlaydate->graphics->drawRoundRect(x,y,width,height,radius,lineWidth,bridgePrimitivePaint(solid,pattern,patterned));}
void bridgeFillRoundedRect(int32_t x,int32_t y,int32_t width,int32_t height,int32_t radius,int32_t solid,uintptr_t pattern,int32_t patterned){activePlaydate->graphics->fillRoundRect(x,y,width,height,radius,bridgePrimitivePaint(solid,pattern,patterned));}
void bridgeSetClipRect(int32_t x,int32_t y,int32_t width,int32_t height){activePlaydate->graphics->setClipRect(x,y,width,height);}
void bridgeClearClipRect(void){activePlaydate->graphics->clearClipRect();}
void bridgeSetDrawOffset(int32_t dx,int32_t dy){activePlaydate->graphics->setDrawOffset(dx,dy);}
void bridgeSetDrawMode(int32_t mode){activePlaydate->graphics->setDrawMode((LCDBitmapDrawMode)mode);}
void bridgeSetLineCapStyle(int32_t style){activePlaydate->graphics->setLineCapStyle((LCDLineCapStyle)style);}
void bridgeSetBackgroundColor(int32_t color){activePlaydate->graphics->setBackgroundColor(color==1?kColorWhite:color==2?kColorBlack:kColorClear);}
void bridgeSetScreenClipRect(int32_t x,int32_t y,int32_t width,int32_t height){activePlaydate->graphics->setScreenClipRect(x,y,width,height);}
void bridgePushContext(uintptr_t bitmap){activePlaydate->graphics->pushContext((LCDBitmap*)bitmap);}
void bridgePopContext(void){activePlaydate->graphics->popContext();}
void bridgeSetStencil(uintptr_t bitmap,int32_t tiled){activePlaydate->graphics->setStencilImage((LCDBitmap*)bitmap,tiled);}
void bridgeClearStencil(void){activePlaydate->graphics->setStencil(NULL);}
uintptr_t bridgeLoadVideo(const char* path){return(uintptr_t)activePlaydate->graphics->video->loadVideo(path);}void bridgeFreeVideo(uintptr_t p){activePlaydate->graphics->video->freePlayer((LCDVideoPlayer*)p);}const char* bridgeVideoError(uintptr_t p){return activePlaydate->graphics->video->getError((LCDVideoPlayer*)p);}void bridgeVideoInfoBits(uintptr_t p,int32_t*w,int32_t*h,uint32_t*r,int32_t*c,int32_t*f){int width,height,count,frame;float rate;activePlaydate->graphics->video->getInfo((LCDVideoPlayer*)p,&width,&height,&rate,&count,&frame);union{float value;uint32_t bits;}v={.value=rate};*w=(int32_t)width;*h=(int32_t)height;*r=v.bits;*c=(int32_t)count;*f=(int32_t)frame;}int32_t bridgeVideoSetContext(uintptr_t p,uintptr_t b){return activePlaydate->graphics->video->setContext((LCDVideoPlayer*)p,(LCDBitmap*)b);}void bridgeVideoUseScreenContext(uintptr_t p){activePlaydate->graphics->video->useScreenContext((LCDVideoPlayer*)p);}int32_t bridgeVideoRenderFrame(uintptr_t p,int32_t f){return activePlaydate->graphics->video->renderFrame((LCDVideoPlayer*)p,(int)f);}
void bridgeDisplaySetRefreshRateBits(uint32_t bits){union{uint32_t bits;float value;}v={.bits=bits};activePlaydate->display->setRefreshRate(v.value);}
int32_t bridgeDisplayWidth(void){return(int32_t)activePlaydate->display->getWidth();}
int32_t bridgeDisplayHeight(void){return(int32_t)activePlaydate->display->getHeight();}
uint32_t bridgeDisplayRefreshRateBits(void){union{float value;uint32_t bits;}v={.value=activePlaydate->display->getRefreshRate()};return v.bits;}
uint32_t bridgeDisplayFPSBits(void){union{float value;uint32_t bits;}v={.value=activePlaydate->display->getFPS()};return v.bits;}
void bridgeDisplaySetInverted(int32_t flag){activePlaydate->display->setInverted(flag);}
void bridgeDisplaySetScale(uint32_t scale){activePlaydate->display->setScale(scale);}
void bridgeDisplaySetMosaic(uint32_t x,uint32_t y){activePlaydate->display->setMosaic(x,y);}
void bridgeDisplaySetFlipped(int32_t x,int32_t y){activePlaydate->display->setFlipped(x,y);}
void bridgeDisplaySetOffset(int32_t x,int32_t y){activePlaydate->display->setOffset(x,y);}
uintptr_t bridgeNewSprite(void) { return (uintptr_t)activePlaydate->sprite->newSprite(); }
extern void goSpriteDraw(uintptr_t,uint32_t,uint32_t,uint32_t,uint32_t,uint32_t,uint32_t,uint32_t,uint32_t);
extern void goSpriteUpdate(uintptr_t); extern int32_t goSpriteCollision(uintptr_t,uintptr_t);
static void bridgeSpriteDraw(LCDSprite*s,PDRect b,PDRect d){goSpriteDraw((uintptr_t)s,bridgeFloatBits(b.x),bridgeFloatBits(b.y),bridgeFloatBits(b.width),bridgeFloatBits(b.height),bridgeFloatBits(d.x),bridgeFloatBits(d.y),bridgeFloatBits(d.width),bridgeFloatBits(d.height));}
static void bridgeSpriteUpdate(LCDSprite*s){goSpriteUpdate((uintptr_t)s);} static SpriteCollisionResponseType bridgeSpriteCollision(LCDSprite*s,LCDSprite*o){return(SpriteCollisionResponseType)goSpriteCollision((uintptr_t)s,(uintptr_t)o);}
void bridgeSpriteSetDrawCallback(uintptr_t s,int32_t v){activePlaydate->sprite->setDrawFunction((LCDSprite*)s,v?bridgeSpriteDraw:NULL);} void bridgeSpriteSetUpdateCallback(uintptr_t s,int32_t v){activePlaydate->sprite->setUpdateFunction((LCDSprite*)s,v?bridgeSpriteUpdate:NULL);} void bridgeSpriteSetCollisionCallback(uintptr_t s,int32_t v){activePlaydate->sprite->setCollisionResponseFunction((LCDSprite*)s,v?bridgeSpriteCollision:NULL);}
void bridgeFreeSprite(uintptr_t sprite) { activePlaydate->sprite->freeSprite((LCDSprite*)sprite); }
void bridgeSpriteSetBitmap(uintptr_t sprite, uintptr_t bitmap) { activePlaydate->sprite->setImage((LCDSprite*)sprite, (LCDBitmap*)bitmap, kBitmapUnflipped); }
void bridgeSpriteSetCenterBits(uintptr_t s,uint32_t x,uint32_t y){union{uint32_t bits;float value;}a={.bits=x},b={.bits=y};activePlaydate->sprite->setCenter((LCDSprite*)s,a.value,b.value);} void bridgeSpriteGetPointBits(uintptr_t s,int32_t center,uint32_t*x,uint32_t*y){union{float value;uint32_t bits;}a,b;if(center)activePlaydate->sprite->getCenter((LCDSprite*)s,&a.value,&b.value);else activePlaydate->sprite->getPosition((LCDSprite*)s,&a.value,&b.value);*x=a.bits;*y=b.bits;}
void bridgeSpriteSetBoundsBits(uintptr_t s,uint32_t x,uint32_t y,uint32_t w,uint32_t h){union{uint32_t bits;float value;}a={.bits=x},b={.bits=y},c={.bits=w},d={.bits=h};activePlaydate->sprite->setBounds((LCDSprite*)s,(PDRect){a.value,b.value,c.value,d.value});} uint32_t bridgeSpriteRectBits(uintptr_t s,int32_t collide,int32_t field){PDRect r=collide?activePlaydate->sprite->getCollideRect((LCDSprite*)s):activePlaydate->sprite->getBounds((LCDSprite*)s);union{float value;uint32_t bits;}v={.value=field==0?r.x:field==1?r.y:field==2?r.width:r.height};return v.bits;}
void bridgeSpriteMoveToBits(uintptr_t sprite, uint32_t x, uint32_t y) { union { uint32_t bits; float value; } px = { .bits = x }, py = { .bits = y }; activePlaydate->sprite->moveTo((LCDSprite*)sprite, px.value, py.value); }
void bridgeSpriteMoveByBits(uintptr_t sprite, uint32_t dx, uint32_t dy) { union { uint32_t bits; float value; } px = { .bits = dx }, py = { .bits = dy }; activePlaydate->sprite->moveBy((LCDSprite*)sprite, px.value, py.value); }
void bridgeSpriteSetVisible(uintptr_t sprite, int32_t visible) { activePlaydate->sprite->setVisible((LCDSprite*)sprite, visible); }
int32_t bridgeSpriteVisible(uintptr_t s){return activePlaydate->sprite->isVisible((LCDSprite*)s);}
void bridgeSpriteSetZIndex(uintptr_t sprite, int32_t z) { activePlaydate->sprite->setZIndex((LCDSprite*)sprite, z); }
int32_t bridgeSpriteZIndex(uintptr_t s){return activePlaydate->sprite->getZIndex((LCDSprite*)s);} void bridgeSpriteSetImageFlip(uintptr_t s,int32_t v){activePlaydate->sprite->setImageFlip((LCDSprite*)s,v);} int32_t bridgeSpriteImageFlip(uintptr_t s){return activePlaydate->sprite->getImageFlip((LCDSprite*)s);} void bridgeSpriteSetDrawMode(uintptr_t s,int32_t v){activePlaydate->sprite->setDrawMode((LCDSprite*)s,v);} void bridgeSpriteSetOpaque(uintptr_t s,int32_t v){activePlaydate->sprite->setOpaque((LCDSprite*)s,v);}
void bridgeSpriteSetStencilImage(uintptr_t s,uintptr_t b,int32_t t){activePlaydate->sprite->setStencilImage((LCDSprite*)s,(LCDBitmap*)b,t);} void bridgeSpriteSetStencilPattern(uintptr_t s,uint8_t*p){activePlaydate->sprite->setStencilPattern((LCDSprite*)s,p);} void bridgeSpriteClearStencil(uintptr_t s){uint8_t open[8]={0xff,0xff,0xff,0xff,0xff,0xff,0xff,0xff};activePlaydate->sprite->setStencilPattern((LCDSprite*)s,open);} void bridgeSpriteSetClipRect(uintptr_t s,int32_t x,int32_t y,int32_t w,int32_t h){activePlaydate->sprite->setClipRect((LCDSprite*)s,LCDMakeRect(x,y,w,h));} void bridgeSpriteClearClipRect(uintptr_t s){activePlaydate->sprite->clearClipRect((LCDSprite*)s);}
void bridgeSpriteSetIgnoresDrawOffset(uintptr_t s,int32_t v){activePlaydate->sprite->setIgnoresDrawOffset((LCDSprite*)s,v);} void bridgeSpriteSetUpdatesEnabled(uintptr_t s,int32_t v){activePlaydate->sprite->setUpdatesEnabled((LCDSprite*)s,v);} int32_t bridgeSpriteUpdatesEnabled(uintptr_t s){return activePlaydate->sprite->updatesEnabled((LCDSprite*)s);} void bridgeSpriteSetCollisionsEnabled(uintptr_t s,int32_t v){activePlaydate->sprite->setCollisionsEnabled((LCDSprite*)s,v);} int32_t bridgeSpriteCollisionsEnabled(uintptr_t s){return activePlaydate->sprite->collisionsEnabled((LCDSprite*)s);}
typedef struct BridgeSpriteTileMap { LCDTileMap* map; uint16_t* tiles; struct BridgeSpriteTileMap* next; } BridgeSpriteTileMap;
static BridgeSpriteTileMap* bridgeSpriteTileMaps;
static BridgeSpriteTileMap* bridgeFindSpriteTileMap(LCDTileMap* map){for(BridgeSpriteTileMap*m=bridgeSpriteTileMaps;m;m=m->next)if(m->map==map)return m;return NULL;}
uintptr_t bridgeNewSpriteTileMap(void){LCDTileMap*map=activePlaydate->graphics->tilemap->newTilemap();if(!map)return 0;BridgeSpriteTileMap*m=activePlaydate->system->realloc(NULL,sizeof(*m));if(!m){activePlaydate->graphics->tilemap->freeTilemap(map);return 0;}m->map=map;m->tiles=NULL;m->next=bridgeSpriteTileMaps;bridgeSpriteTileMaps=m;return(uintptr_t)m;}
void bridgeFreeSpriteTileMap(uintptr_t h){BridgeSpriteTileMap*m=(BridgeSpriteTileMap*)h;BridgeSpriteTileMap**p=&bridgeSpriteTileMaps;while(*p&&*p!=m)p=&(*p)->next;if(*p)*p=m->next;if(m->tiles)activePlaydate->system->realloc(m->tiles,0);activePlaydate->graphics->tilemap->freeTilemap(m->map);activePlaydate->system->realloc(m,0);}
void bridgeSpriteTileMapSetImageTable(uintptr_t h,uintptr_t t){activePlaydate->graphics->tilemap->setImageTable(((BridgeSpriteTileMap*)h)->map,(LCDBitmapTable*)t);} void bridgeSpriteTileMapSetSize(uintptr_t h,int32_t w,int32_t z){activePlaydate->graphics->tilemap->setSize(((BridgeSpriteTileMap*)h)->map,w,z);} void bridgeSpriteTileMapSize(uintptr_t h,int32_t*w,int32_t*z){activePlaydate->graphics->tilemap->getSize(((BridgeSpriteTileMap*)h)->map,(int*)w,(int*)z);} void bridgeSpriteTileMapPixelSize(uintptr_t h,uint32_t*w,uint32_t*z){activePlaydate->graphics->tilemap->getPixelSize(((BridgeSpriteTileMap*)h)->map,w,z);}
void bridgeSpriteTileMapSetTiles(uintptr_t h,uint16_t*v,int32_t count,int32_t row){BridgeSpriteTileMap*m=(BridgeSpriteTileMap*)h;uint16_t*copy=activePlaydate->system->realloc(m->tiles,(size_t)count*sizeof(uint16_t));if(!copy)return;memcpy(copy,v,(size_t)count*sizeof(uint16_t));m->tiles=copy;activePlaydate->graphics->tilemap->setTiles(m->map,m->tiles,count,row);}
void bridgeSpriteTileMapSetTile(uintptr_t h,int32_t x,int32_t y,uint16_t v){activePlaydate->graphics->tilemap->setTileAtPosition(((BridgeSpriteTileMap*)h)->map,x,y,v);} int32_t bridgeSpriteTileMapTile(uintptr_t h,int32_t x,int32_t y){return activePlaydate->graphics->tilemap->getTileAtPosition(((BridgeSpriteTileMap*)h)->map,x,y);} void bridgeSpriteSetTileMap(uintptr_t s,uintptr_t h){activePlaydate->sprite->setTilemap((LCDSprite*)s,h?((BridgeSpriteTileMap*)h)->map:NULL);} uintptr_t bridgeSpriteTileMap(uintptr_t s){BridgeSpriteTileMap*m=bridgeFindSpriteTileMap(activePlaydate->sprite->getTilemap((LCDSprite*)s));return(uintptr_t)m;}
void bridgeSpriteSetCollideRectBits(uintptr_t sprite, uint32_t x, uint32_t y, uint32_t width, uint32_t height) { union { uint32_t bits; float value; } px={.bits=x}, py={.bits=y}, pw={.bits=width}, ph={.bits=height}; activePlaydate->sprite->setCollideRect((LCDSprite*)sprite, (PDRect){px.value,py.value,pw.value,ph.value}); }
void bridgeSpriteClearCollideRect(uintptr_t sprite) { activePlaydate->sprite->clearCollideRect((LCDSprite*)sprite); }
void bridgeSpriteSetTag(uintptr_t sprite, uint8_t tag) { activePlaydate->sprite->setTag((LCDSprite*)sprite, tag); }
uint8_t bridgeSpriteTag(uintptr_t s){return activePlaydate->sprite->getTag((LCDSprite*)s);}
void bridgeSpriteMarkDirty(uintptr_t sprite){activePlaydate->sprite->markDirty((LCDSprite*)sprite);}
void bridgeSpriteMarkDirtyRectBits(uintptr_t sprite,uint32_t x,uint32_t y,uint32_t w,uint32_t h){union{uint32_t bits;float value;}a={.bits=x},b={.bits=y},c={.bits=w},d={.bits=h};activePlaydate->sprite->markDirtyRect((LCDSprite*)sprite,(PDRect){a.value,b.value,c.value,d.value});}
uintptr_t bridgeSpriteMoveWithCollisionsBits(uintptr_t sprite, uint32_t x, uint32_t y, uint32_t* actualX, uint32_t* actualY, int32_t* count) { union { uint32_t bits; float value; } px={.bits=x}, py={.bits=y}, ax, ay; SpriteCollisionInfo* result=activePlaydate->sprite->moveWithCollisions((LCDSprite*)sprite,px.value,py.value,&ax.value,&ay.value,(int*)count); *actualX=ax.bits; *actualY=ay.bits; return (uintptr_t)result; }
uintptr_t bridgeSpriteCheckCollisionsBits(uintptr_t sprite,uint32_t x,uint32_t y,uint32_t*ax,uint32_t*ay,int32_t*n){union{uint32_t bits;float value;}px={.bits=x},py={.bits=y},a,b;SpriteCollisionInfo*r=activePlaydate->sprite->checkCollisions((LCDSprite*)sprite,px.value,py.value,&a.value,&b.value,(int*)n);*ax=a.bits;*ay=b.bits;return(uintptr_t)r;}
uint32_t bridgeCollisionValueBits(uintptr_t collisions, int32_t index, int32_t field) { SpriteCollisionInfo* c=&((SpriteCollisionInfo*)collisions)[index]; union { float value; uint32_t bits; } v; switch(field){case 0:v.value=c->ti;break;case 1:v.value=c->move.x;break;case 2:v.value=c->move.y;break;case 3:v.value=c->normal.x;break;case 4:v.value=c->normal.y;break;case 5:v.value=c->touch.x;break;case 6:v.value=c->touch.y;break;case 7:v.value=c->spriteRect.x;break;case 8:v.value=c->spriteRect.y;break;case 9:v.value=c->spriteRect.width;break;case 10:v.value=c->spriteRect.height;break;case 11:v.value=c->otherRect.x;break;case 12:v.value=c->otherRect.y;break;case 13:v.value=c->otherRect.width;break;case 16:v.value=c->otherRect.height;break;case 14:return (uint32_t)c->responseType;case 15:return (uint32_t)c->overlaps;default:return 0;} return v.bits; }
uintptr_t bridgeCollisionOther(uintptr_t collisions, int32_t index) { return (uintptr_t)((SpriteCollisionInfo*)collisions)[index].other; }
void bridgeFreeList(uintptr_t list) { if(list) activePlaydate->system->realloc((void*)list,0); }
uintptr_t bridgeQuerySpritesAtPointBits(uint32_t x,uint32_t y,int32_t* count) { union {uint32_t bits;float value;} px={.bits=x},py={.bits=y}; return (uintptr_t)activePlaydate->sprite->querySpritesAtPoint(px.value,py.value,(int*)count); }
uintptr_t bridgeQuerySpritesInRectBits(uint32_t x,uint32_t y,uint32_t width,uint32_t height,int32_t* count) { union {uint32_t bits;float value;} px={.bits=x},py={.bits=y},pw={.bits=width},ph={.bits=height}; return (uintptr_t)activePlaydate->sprite->querySpritesInRect(px.value,py.value,pw.value,ph.value,(int*)count); }
uintptr_t bridgeQuerySpritesAlongLineBits(uint32_t a,uint32_t b,uint32_t c,uint32_t d,int32_t*n){union{uint32_t bits;float value;}x1={.bits=a},y1={.bits=b},x2={.bits=c},y2={.bits=d};return(uintptr_t)activePlaydate->sprite->querySpritesAlongLine(x1.value,y1.value,x2.value,y2.value,(int*)n);} uintptr_t bridgeQuerySpriteInfoAlongLineBits(uint32_t a,uint32_t b,uint32_t c,uint32_t d,int32_t*n){union{uint32_t bits;float value;}x1={.bits=a},y1={.bits=b},x2={.bits=c},y2={.bits=d};return(uintptr_t)activePlaydate->sprite->querySpriteInfoAlongLine(x1.value,y1.value,x2.value,y2.value,(int*)n);}
uint32_t bridgeSpriteQueryInfoValueBits(uintptr_t p,int32_t i,int32_t f){SpriteQueryInfo*q=&((SpriteQueryInfo*)p)[i];union{float value;uint32_t bits;}v;switch(f){case 0:v.value=q->ti1;break;case 1:v.value=q->ti2;break;case 2:v.value=q->entryPoint.x;break;case 3:v.value=q->entryPoint.y;break;case 4:v.value=q->exitPoint.x;break;default:v.value=q->exitPoint.y;}return v.bits;} uintptr_t bridgeSpriteQueryInfoSprite(uintptr_t p,int32_t i){return(uintptr_t)((SpriteQueryInfo*)p)[i].sprite;}
uintptr_t bridgeSpriteListItem(uintptr_t list,int32_t index) { return (uintptr_t)((LCDSprite**)list)[index]; }
uintptr_t bridgeOverlappingSprites(uintptr_t sprite,int32_t* count) { return (uintptr_t)activePlaydate->sprite->overlappingSprites((LCDSprite*)sprite,(int*)count); }
void bridgeSpriteAdd(uintptr_t sprite) { activePlaydate->sprite->addSprite((LCDSprite*)sprite); }
void bridgeSpriteRemove(uintptr_t sprite) { activePlaydate->sprite->removeSprite((LCDSprite*)sprite); }
void bridgeSpriteRemoveMany(uintptr_t*s,int32_t n){activePlaydate->sprite->removeSprites((LCDSprite**)s,n);} void bridgeRemoveAllSprites(void){activePlaydate->sprite->removeAllSprites();} int32_t bridgeSpriteCount(void){return activePlaydate->sprite->getSpriteCount();} void bridgeResetCollisionWorld(void){activePlaydate->sprite->resetCollisionWorld();}
void bridgeSetAlwaysRedraw(int32_t flag){activePlaydate->sprite->setAlwaysRedraw(flag);}
void bridgeAddDirtyRect(int32_t x,int32_t y,int32_t w,int32_t h){activePlaydate->sprite->addDirtyRect(LCDMakeRect(x,y,w,h));}
void bridgeUpdateAndDrawSprites(void) { activePlaydate->sprite->updateAndDrawSprites(); }
typedef struct { AudioSample* sample; SamplePlayer* player; int ownsSample; } BridgeSoundEffect;
static void bridgeAudioFinishCallback(SoundSource* source,void* userdata);
uintptr_t bridgeLoadSoundEffect(const char* path) { AudioSample* sample=activePlaydate->sound->sample->load(path); if(!sample)return 0; SamplePlayer* player=activePlaydate->sound->sampleplayer->newPlayer(); if(!player){activePlaydate->sound->sample->freeSample(sample);return 0;} BridgeSoundEffect* effect=activePlaydate->system->realloc(NULL,sizeof(BridgeSoundEffect)); if(!effect){activePlaydate->sound->sampleplayer->freePlayer(player);activePlaydate->sound->sample->freeSample(sample);return 0;} effect->sample=sample;effect->player=player;effect->ownsSample=1;activePlaydate->sound->sampleplayer->setSample(player,sample);return(uintptr_t)effect; }
uintptr_t bridgeNewPCMPlayer(const int16_t*samples,int32_t count,uint32_t rate){if(!samples||count<=0||rate==0)return 0;int32_t bytes=count*2;uint8_t*copy=activePlaydate->system->realloc(NULL,(size_t)bytes);if(!copy)return 0;memcpy(copy,samples,(size_t)bytes);AudioSample*sample=activePlaydate->sound->sample->newSampleFromData(copy,kSound16bitMono,rate,bytes,1);if(!sample){activePlaydate->system->realloc(copy,0);return 0;}SamplePlayer*player=activePlaydate->sound->sampleplayer->newPlayer();if(!player){activePlaydate->sound->sample->freeSample(sample);return 0;}BridgeSoundEffect*effect=activePlaydate->system->realloc(NULL,sizeof(BridgeSoundEffect));if(!effect){activePlaydate->sound->sampleplayer->freePlayer(player);activePlaydate->sound->sample->freeSample(sample);return 0;}effect->sample=sample;effect->player=player;effect->ownsSample=1;activePlaydate->sound->sampleplayer->setSample(player,sample);return(uintptr_t)effect;}
uintptr_t bridgeNewSample(int32_t n){return(uintptr_t)activePlaydate->sound->sample->newSampleBuffer(n);} uintptr_t bridgeLoadSample(const char*p){return(uintptr_t)activePlaydate->sound->sample->load(p);} uintptr_t bridgeNewSampleFromData(const uint8_t*d,int32_t n,int32_t f,uint32_t r){uint8_t*c=activePlaydate->system->realloc(NULL,n);if(!c)return 0;memcpy(c,d,n);AudioSample*s=activePlaydate->sound->sample->newSampleFromData(c,(SoundFormat)f,r,n,1);if(!s)activePlaydate->system->realloc(c,0);return(uintptr_t)s;} int32_t bridgeSampleLoad(uintptr_t s,const char*p){return activePlaydate->sound->sample->loadIntoSample((AudioSample*)s,p);} void bridgeSampleData(uintptr_t s,uintptr_t*d,int32_t*f,uint32_t*r,uint32_t*n){uint8_t*p;SoundFormat v;activePlaydate->sound->sample->getData((AudioSample*)s,&p,&v,r,n);*d=(uintptr_t)p;*f=(int32_t)v;} uint32_t bridgeSampleLengthBits(uintptr_t s){union{float f;uint32_t u;}v={.f=activePlaydate->sound->sample->getLength((AudioSample*)s)};return v.u;} int32_t bridgeSampleDecompress(uintptr_t s){return activePlaydate->sound->sample->decompress((AudioSample*)s);} void bridgeFreeSample(uintptr_t s){activePlaydate->sound->sample->freeSample((AudioSample*)s);} uintptr_t bridgeNewEmptySamplePlayer(void){SamplePlayer*p=activePlaydate->sound->sampleplayer->newPlayer();if(!p)return 0;BridgeSoundEffect*e=activePlaydate->system->realloc(NULL,sizeof(BridgeSoundEffect));if(!e){activePlaydate->sound->sampleplayer->freePlayer(p);return 0;}e->sample=NULL;e->player=p;e->ownsSample=0;return(uintptr_t)e;}
static BridgeSoundEffect* bridgeEffect(uintptr_t effect){return(BridgeSoundEffect*)effect;}
int32_t bridgeSoundEffectPlay(uintptr_t effect){return activePlaydate->sound->sampleplayer->play(bridgeEffect(effect)->player,1,1.0f);}
int32_t bridgeSamplePlayerPlayBits(uintptr_t effect,int32_t repeat,uint32_t rate){union{uint32_t bits;float value;}r={.bits=rate};return activePlaydate->sound->sampleplayer->play(bridgeEffect(effect)->player,repeat,r.value);}
void bridgeSoundEffectStop(uintptr_t effect){activePlaydate->sound->sampleplayer->stop(bridgeEffect(effect)->player);}
void bridgeSoundEffectSetVolumeBits(uintptr_t effect,uint32_t left,uint32_t right){union{uint32_t bits;float value;}l={.bits=left},r={.bits=right};activePlaydate->sound->sampleplayer->setVolume(bridgeEffect(effect)->player,l.value,r.value);}
void bridgeSoundEffectVolumeBits(uintptr_t effect,uint32_t* left,uint32_t* right){union{float value;uint32_t bits;}l,r;activePlaydate->sound->sampleplayer->getVolume(bridgeEffect(effect)->player,&l.value,&r.value);*left=l.bits;*right=r.bits;}
int32_t bridgeSoundEffectIsPlaying(uintptr_t effect){return activePlaydate->sound->sampleplayer->isPlaying(bridgeEffect(effect)->player);}
void bridgeSoundEffectPause(uintptr_t effect,int32_t paused){activePlaydate->sound->sampleplayer->setPaused(bridgeEffect(effect)->player,paused);}
uint32_t bridgeSamplePlayerLengthBits(uintptr_t effect){union{float value;uint32_t bits;}v={.value=activePlaydate->sound->sampleplayer->getLength(bridgeEffect(effect)->player)};return v.bits;}
void bridgeSamplePlayerSetOffsetBits(uintptr_t effect,uint32_t offset){union{uint32_t bits;float value;}v={.bits=offset};activePlaydate->sound->sampleplayer->setOffset(bridgeEffect(effect)->player,v.value);}
uint32_t bridgeSamplePlayerOffsetBits(uintptr_t effect){union{float value;uint32_t bits;}v={.value=activePlaydate->sound->sampleplayer->getOffset(bridgeEffect(effect)->player)};return v.bits;}
void bridgeSamplePlayerSetRateBits(uintptr_t effect,uint32_t rate){union{uint32_t bits;float value;}v={.bits=rate};activePlaydate->sound->sampleplayer->setRate(bridgeEffect(effect)->player,v.value);}
void bridgeSamplePlayerSetRateModulator(uintptr_t e,uintptr_t s){activePlaydate->sound->sampleplayer->setRateModulator(bridgeEffect(e)->player,(PDSynthSignalValue*)s);}
uint32_t bridgeSamplePlayerRateBits(uintptr_t effect){union{float value;uint32_t bits;}v={.value=activePlaydate->sound->sampleplayer->getRate(bridgeEffect(effect)->player)};return v.bits;}
void bridgeSoundEffectSetLoopCallback(uintptr_t e,uint32_t c){activePlaydate->sound->sampleplayer->setLoopCallback(bridgeEffect(e)->player,c?bridgeAudioFinishCallback:NULL,(void*)(uintptr_t)c);} void bridgeSamplePlayerSetSample(uintptr_t e,uintptr_t s){BridgeSoundEffect*v=bridgeEffect(e);if(v->ownsSample&&v->sample)activePlaydate->sound->sample->freeSample(v->sample);v->sample=(AudioSample*)s;v->ownsSample=0;activePlaydate->sound->sampleplayer->setSample(v->player,v->sample);} void bridgeSamplePlayerSetPlayRange(uintptr_t e,int32_t s,int32_t n){activePlaydate->sound->sampleplayer->setPlayRange(bridgeEffect(e)->player,s,n);}
void bridgeFreeSoundEffect(uintptr_t effect){BridgeSoundEffect* value=bridgeEffect(effect);activePlaydate->sound->sampleplayer->freePlayer(value->player);if(value->ownsSample&&value->sample)activePlaydate->sound->sample->freeSample(value->sample);activePlaydate->system->realloc(value,0);}
uintptr_t bridgeLoadFilePlayer(const char* path){FilePlayer* player=activePlaydate->sound->fileplayer->newPlayer();if(!player)return 0;if(!activePlaydate->sound->fileplayer->loadIntoPlayer(player,path)){activePlaydate->sound->fileplayer->freePlayer(player);return 0;}return(uintptr_t)player;}
int32_t bridgeFilePlayerPlay(uintptr_t player){return activePlaydate->sound->fileplayer->play((FilePlayer*)player,1);}
void bridgeFilePlayerStop(uintptr_t player){activePlaydate->sound->fileplayer->stop((FilePlayer*)player);}
void bridgeFilePlayerSetVolumeBits(uintptr_t player,uint32_t left,uint32_t right){union{uint32_t bits;float value;}l={.bits=left},r={.bits=right};activePlaydate->sound->fileplayer->setVolume((FilePlayer*)player,l.value,r.value);}
void bridgeFilePlayerVolumeBits(uintptr_t player,uint32_t* left,uint32_t* right){union{float value;uint32_t bits;}l,r;activePlaydate->sound->fileplayer->getVolume((FilePlayer*)player,&l.value,&r.value);*left=l.bits;*right=r.bits;}
int32_t bridgeFilePlayerIsPlaying(uintptr_t player){return activePlaydate->sound->fileplayer->isPlaying((FilePlayer*)player);}
void bridgeFilePlayerPause(uintptr_t player){activePlaydate->sound->fileplayer->pause((FilePlayer*)player);}
void bridgeFilePlayerSetRateBits(uintptr_t player,uint32_t rate){union{uint32_t bits;float value;}v={.bits=rate};activePlaydate->sound->fileplayer->setRate((FilePlayer*)player,v.value);}
void bridgeFilePlayerSetRateModulator(uintptr_t p,uintptr_t s){activePlaydate->sound->fileplayer->setRateModulator((FilePlayer*)p,(PDSynthSignalValue*)s);}
uint32_t bridgeFilePlayerRateBits(uintptr_t player){union{float value;uint32_t bits;}v={.value=activePlaydate->sound->fileplayer->getRate((FilePlayer*)player)};return v.bits;}
uint32_t bridgeCurrentAudioTime(void){return activePlaydate->sound->getCurrentTime();}
extern void goAudioCallback(uint32_t callback,int32_t oneShot);
extern void goMicrophonePermission(int32_t allowed);
extern int32_t goMicrophoneSamples(uintptr_t data,int32_t length);
typedef struct{uint32_t callback;int32_t oneShot;}BridgePendingAudioCallback;
static volatile BridgePendingAudioCallback bridgeAudioCallbacks[8];static volatile uint32_t bridgeAudioCallbackRead=0,bridgeAudioCallbackWrite=0;
static void bridgeQueueAudioCallback(uint32_t callback,int32_t oneShot){uint32_t next=(bridgeAudioCallbackWrite+1)%8;if(next==bridgeAudioCallbackRead)return;bridgeAudioCallbacks[bridgeAudioCallbackWrite].callback=callback;bridgeAudioCallbacks[bridgeAudioCallbackWrite].oneShot=oneShot;bridgeAudioCallbackWrite=next;}
int32_t bridgePollAudioCallback(uint32_t* callback,int32_t* oneShot){if(bridgeAudioCallbackRead==bridgeAudioCallbackWrite)return 0;*callback=bridgeAudioCallbacks[bridgeAudioCallbackRead].callback;*oneShot=bridgeAudioCallbacks[bridgeAudioCallbackRead].oneShot;bridgeAudioCallbackRead=(bridgeAudioCallbackRead+1)%8;return 1;}
static void bridgeAudioFinishCallback(SoundSource* source,void* userdata){(void)source;bridgeQueueAudioCallback((uint32_t)(uintptr_t)userdata,0);}
static void bridgeAudioFadeCallback(SoundSource* source,void* userdata){(void)source;bridgeQueueAudioCallback((uint32_t)(uintptr_t)userdata,1);}
void bridgeSoundEffectSetFinishCallback(uintptr_t effect,uint32_t callback){activePlaydate->sound->sampleplayer->setFinishCallback(bridgeEffect(effect)->player,callback?bridgeAudioFinishCallback:NULL,(void*)(uintptr_t)callback);}
void bridgeFilePlayerSetFinishCallback(uintptr_t player,uint32_t callback){activePlaydate->sound->fileplayer->setFinishCallback((FilePlayer*)player,callback?bridgeAudioFinishCallback:NULL,(void*)(uintptr_t)callback);}
int32_t bridgeFilePlayerLoad(uintptr_t p,const char*s){return activePlaydate->sound->fileplayer->loadIntoPlayer((FilePlayer*)p,s);} void bridgeFilePlayerSetBufferLengthBits(uintptr_t p,uint32_t n){union{uint32_t u;float f;}v={.u=n};activePlaydate->sound->fileplayer->setBufferLength((FilePlayer*)p,v.f);} void bridgeFilePlayerSetLoopRangeBits(uintptr_t p,uint32_t s,uint32_t e){union{uint32_t u;float f;}a={.u=s},b={.u=e};activePlaydate->sound->fileplayer->setLoopRange((FilePlayer*)p,a.f,b.f);} int32_t bridgeFilePlayerDidUnderrun(uintptr_t p){return activePlaydate->sound->fileplayer->didUnderrun((FilePlayer*)p);} void bridgeFilePlayerSetStopOnUnderrun(uintptr_t p,int32_t v){activePlaydate->sound->fileplayer->setStopOnUnderrun((FilePlayer*)p,v);} void bridgeFilePlayerSetLoopCallback(uintptr_t p,uint32_t c){activePlaydate->sound->fileplayer->setLoopCallback((FilePlayer*)p,c?bridgeAudioFinishCallback:NULL,(void*)(uintptr_t)c);}
void bridgeFilePlayerFadeVolumeBits(uintptr_t player,uint32_t left,uint32_t right,uint32_t frames,uint32_t callback){union{uint32_t bits;float value;}l={.bits=left},r={.bits=right};activePlaydate->sound->fileplayer->fadeVolume((FilePlayer*)player,l.value,r.value,(int32_t)frames,callback?bridgeAudioFadeCallback:NULL,(void*)(uintptr_t)callback);}
void bridgeFreeFilePlayer(uintptr_t player){activePlaydate->sound->fileplayer->freePlayer((FilePlayer*)player);}
