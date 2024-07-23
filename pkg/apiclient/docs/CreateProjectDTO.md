# CreateProjectDTO

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ExistingProjectConfig** | Pointer to [**ExistingProjectConfigDTO**](ExistingProjectConfigDTO.md) |  | [optional] 
**NewProjectConfig** | Pointer to [**CreateProjectConfigDTO**](CreateProjectConfigDTO.md) |  | [optional] 

## Methods

### NewCreateProjectDTO

`func NewCreateProjectDTO() *CreateProjectDTO`

NewCreateProjectDTO instantiates a new CreateProjectDTO object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateProjectDTOWithDefaults

`func NewCreateProjectDTOWithDefaults() *CreateProjectDTO`

NewCreateProjectDTOWithDefaults instantiates a new CreateProjectDTO object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetExistingProjectConfig

`func (o *CreateProjectDTO) GetExistingProjectConfig() ExistingProjectConfigDTO`

GetExistingProjectConfig returns the ExistingProjectConfig field if non-nil, zero value otherwise.

### GetExistingProjectConfigOk

`func (o *CreateProjectDTO) GetExistingProjectConfigOk() (*ExistingProjectConfigDTO, bool)`

GetExistingProjectConfigOk returns a tuple with the ExistingProjectConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExistingProjectConfig

`func (o *CreateProjectDTO) SetExistingProjectConfig(v ExistingProjectConfigDTO)`

SetExistingProjectConfig sets ExistingProjectConfig field to given value.

### HasExistingProjectConfig

`func (o *CreateProjectDTO) HasExistingProjectConfig() bool`

HasExistingProjectConfig returns a boolean if a field has been set.

### GetNewProjectConfig

`func (o *CreateProjectDTO) GetNewProjectConfig() CreateProjectConfigDTO`

GetNewProjectConfig returns the NewProjectConfig field if non-nil, zero value otherwise.

### GetNewProjectConfigOk

`func (o *CreateProjectDTO) GetNewProjectConfigOk() (*CreateProjectConfigDTO, bool)`

GetNewProjectConfigOk returns a tuple with the NewProjectConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNewProjectConfig

`func (o *CreateProjectDTO) SetNewProjectConfig(v CreateProjectConfigDTO)`

SetNewProjectConfig sets NewProjectConfig field to given value.

### HasNewProjectConfig

`func (o *CreateProjectDTO) HasNewProjectConfig() bool`

HasNewProjectConfig returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


